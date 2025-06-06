package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Enhanced Invoice struct with print layout support
type Invoice struct {
    InvoiceID       string       `json:"invoice_id"`
    TransactionID   string       `json:"transaction_id"`
    CustomerName    string       `json:"customer_name"`
    CustomerPhone   string       `json:"customer_phone"`
    InvoiceDate     time.Time    `json:"invoice_date"`
    DueDate         time.Time    `json:"due_date"`
    Items           []InvoiceItem `json:"items"`
    Subtotal        float64      `json:"subtotal"`
    Tax             float64      `json:"tax"`
    Discount        float64      `json:"discount"`
    Total           float64      `json:"total"`
    PaymentStatus   string       `json:"payment_status"`
    PaymentMethod   string       `json:"payment_method"`
    Notes           string       `json:"notes"`
    BusinessInfo    BusinessInfo `json:"business_info"`
    CashierName     string       `json:"cashier_name"`
    InvoiceNumber   int          `json:"invoice_number"`
}

type InvoiceItem struct {
    ProductName string  `json:"product_name"`
    Quantity    int     `json:"quantity"`
    UnitPrice   float64 `json:"unit_price"`
    TotalPrice  float64 `json:"total_price"`
    Description string  `json:"description"`
    Category    string  `json:"category"`
}

type BusinessInfo struct {
    Name        string `json:"name"`
    Address     string `json:"address"`
    Phone       string `json:"phone"`
    Email       string `json:"email"`
    TaxID       string `json:"tax_id"`
    Logo        string `json:"logo"`
    Website     string `json:"website"`
    BankAccount string `json:"bank_account"`
}

// Get invoice by transaction ID
func GetInvoiceByTransactionID(w http.ResponseWriter, r *http.Request) {
    // Remove enableCors(&w) - handled by main.go now
    
    transactionID := r.URL.Query().Get("transaction_id")
    if transactionID == "" {
        http.Error(w, "Transaction ID is required", http.StatusBadRequest)
        return
    }

    db := Koneksi()
    defer db.Close()

    // Get transaction details
    sqlQuery := `
        SELECT 
            td.transaksi_id, 
            td.nama_produk, 
            td.harga AS harga_jual, 
            p.harga_beli, 
            td.jumlah_terjual, 
            td.total_harga, 
            th.tanggal_transaksi,
            th.total_item
        FROM transaksi_detail td
        INNER JOIN produk p ON td.nama_produk = p.nama
        INNER JOIN transaksi_header th ON td.transaksi_id = th.transaksi_id
        WHERE td.transaksi_id = $1
        ORDER BY td.nama_produk
    `

    rows, err := db.Query(sqlQuery, transactionID)
    if err != nil {
        http.Error(w, "Failed to fetch transaction details: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var invoice Invoice
    var items []InvoiceItem
    var subtotal float64

    // Business information
    invoice.BusinessInfo = BusinessInfo{
        Name:    "Bintang Jaya Frozen Food",
        Address: "Jl. Raya Frozen No. 123, Jakarta Selatan",
        Phone:   "+62 21 1234 5678",
        Email:   "info@bintangjaya.com",
        TaxID:   "01.234.567.8-901.000",
        Logo:    "/assets/logo.png",
    }

    for rows.Next() {
        var transID string
        var productName string
        var hargaJual float64
        var hargaBeli float64
        var quantity int
        var totalPrice float64
        var transactionDate time.Time
        var totalItems int

        err := rows.Scan(&transID, &productName, &hargaJual, &hargaBeli, &quantity, &totalPrice, &transactionDate, &totalItems)
        if err != nil {
            http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Set invoice header info (only once)
        if invoice.TransactionID == "" {
            invoice.InvoiceID = "INV-" + transID + "-" + time.Now().Format("20060102")
            invoice.TransactionID = transID
            invoice.InvoiceDate = transactionDate
            invoice.DueDate = transactionDate.AddDate(0, 0, 30)
            invoice.PaymentStatus = "PAID"
            invoice.PaymentMethod = "CASH"
        }

        // Add item to invoice
        item := InvoiceItem{
            ProductName: productName,
            Quantity:    quantity,
            UnitPrice:   hargaJual,
            TotalPrice:  totalPrice,
            Description: fmt.Sprintf("Frozen food item - %s", productName),
        }
        items = append(items, item)
        subtotal += totalPrice
    }

    if len(items) == 0 {
        http.Error(w, "Transaction not found", http.StatusNotFound)
        return
    }

    invoice.Items = items
    invoice.Subtotal = subtotal
    invoice.Tax = subtotal * 0.11
    invoice.Discount = 0.0
    invoice.Total = invoice.Subtotal + invoice.Tax - invoice.Discount

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(invoice)
}

// Get all invoices with pagination and filters
func GetAllInvoices(w http.ResponseWriter, r *http.Request) {
    db := Koneksi()
    defer db.Close()

    // Get query parameters
    page := r.URL.Query().Get("page")
    if page == "" {
        page = "1"
    }
    
    limit := r.URL.Query().Get("limit")
    if limit == "" {
        limit = "10"
    }

    startDate := r.URL.Query().Get("start_date")
    endDate := r.URL.Query().Get("end_date")

    // Build the query with filters
    whereClause := "WHERE 1=1"
    var args []interface{}
    argCount := 0

    if startDate != "" {
        argCount++
        whereClause += fmt.Sprintf(" AND DATE(th.tanggal_transaksi) >= $%d", argCount)
        args = append(args, startDate)
    }

    if endDate != "" {
        argCount++
        whereClause += fmt.Sprintf(" AND DATE(th.tanggal_transaksi) <= $%d", argCount)
        args = append(args, endDate)
    }

    // ✅ FIX: Get total count first (without LIMIT)
    countQuery := fmt.Sprintf(`
        SELECT COUNT(DISTINCT th.transaksi_id)
        FROM transaksi_header th
        INNER JOIN transaksi_detail td ON th.transaksi_id = td.transaksi_id
        %s
    `, whereClause)

    var totalCount int
    err := db.QueryRow(countQuery, args...).Scan(&totalCount)
    if err != nil {
        http.Error(w, "Failed to get total count: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // ✅ FIX: Get paginated data
    sqlQuery := fmt.Sprintf(`
        SELECT 
            th.transaksi_id,
            th.tanggal_transaksi,
            th.total_item,
            SUM(td.total_harga) as total_amount,
            COUNT(td.nama_produk) as item_count
        FROM transaksi_header th
        INNER JOIN transaksi_detail td ON th.transaksi_id = td.transaksi_id
        %s
        GROUP BY th.transaksi_id, th.tanggal_transaksi, th.total_item
        ORDER BY th.tanggal_transaksi DESC
        LIMIT %s OFFSET %s
    `, whereClause, limit, fmt.Sprintf("%d", (getIntOrDefault(page, 1)-1)*getIntOrDefault(limit, 10)))

    rows, err := db.Query(sqlQuery, args...)
    if err != nil {
        http.Error(w, "Failed to fetch invoices: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var invoices []map[string]interface{}

    for rows.Next() {
        var transactionID string
        var transactionDate time.Time
        var totalItems int
        var totalAmount float64
        var itemCount int

        err := rows.Scan(&transactionID, &transactionDate, &totalItems, &totalAmount, &itemCount)
        if err != nil {
            http.Error(w, "Failed to scan invoice row: "+err.Error(), http.StatusInternalServerError)
            return
        }

        invoice := map[string]interface{}{
            "invoice_id":     "INV-" + transactionID + "-" + transactionDate.Format("20060102"),
            "transaction_id": transactionID,
            "invoice_date":   transactionDate,
            "total_amount":   totalAmount + (totalAmount * 0.11),
            "item_count":     itemCount,
            "payment_status": "PAID",
            "payment_method": "CASH",
        }
        invoices = append(invoices, invoice)
    }

    // ✅ FIX: Return both paginated data and total count
    response := map[string]interface{}{
        "invoices":    invoices,
        "page":        getIntOrDefault(page, 1),
        "limit":       getIntOrDefault(limit, 10),
        "total":       len(invoices),        // Current page count
        "total_count": totalCount,           // ✅ Total count of ALL invoices
        "total_pages": (totalCount + getIntOrDefault(limit, 10) - 1) / getIntOrDefault(limit, 10),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Generate HTML invoice for printing/PDF
func GenerateInvoiceHTML(w http.ResponseWriter, r *http.Request) {
    // Add logging for debugging
    fmt.Printf("GenerateInvoiceHTML called with method: %s\n", r.Method)
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    transactionID := r.URL.Query().Get("transaction_id")
    if transactionID == "" {
        fmt.Printf("Error: Transaction ID is missing\n")
        http.Error(w, "Transaction ID is required", http.StatusBadRequest)
        return
    }

    fmt.Printf("Processing transaction ID: %s\n", transactionID)

    // Get invoice data with error handling
    invoice, err := getInvoiceData(transactionID)
    if err != nil {
        fmt.Printf("Error getting invoice data: %v\n", err)
        http.Error(w, fmt.Sprintf("Failed to get invoice data: %v", err), http.StatusInternalServerError)
        return
    }

    fmt.Printf("Invoice data retrieved successfully\n")

    // Generate HTML template with error handling
    htmlContent := generateInvoiceHTMLTemplate(invoice)
    if htmlContent == "" {
        fmt.Printf("Error: HTML content is empty\n")
        http.Error(w, "Failed to generate HTML content", http.StatusInternalServerError)
        return
    }

    // Check if HTML contains error messages
    if strings.Contains(htmlContent, "Error parsing template") || strings.Contains(htmlContent, "Error executing template") {
        fmt.Printf("Template error detected: %s\n", htmlContent)
        http.Error(w, htmlContent, http.StatusInternalServerError)
        return
    }

    fmt.Printf("HTML content generated successfully, length: %d\n", len(htmlContent))

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(htmlContent))
}

// Auto-generate invoice after transaction
func AutoGenerateInvoice(w http.ResponseWriter, r *http.Request) {
    // Remove enableCors(&w) - handled by main.go now
    
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        TransactionID string `json:"transaction_id"`
        CustomerName  string `json:"customer_name"`
        CustomerPhone string `json:"customer_phone"`
        CashierName   string `json:"cashier_name"`
        AutoPrint     bool   `json:"auto_print"`
    }

    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate invoice HTML
    invoice, err := getInvoiceData(request.TransactionID)
    if err != nil {
        http.Error(w, "Failed to generate invoice: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Update customer info if provided
    if request.CustomerName != "" {
        invoice.CustomerName = request.CustomerName
    }
    if request.CustomerPhone != "" {
        invoice.CustomerPhone = request.CustomerPhone
    }
    if request.CashierName != "" {
        invoice.CashierName = request.CashierName
    }

    htmlContent := generateInvoiceHTMLTemplate(invoice)

    response := map[string]interface{}{
        "success":        true,
        "invoice_id":     invoice.InvoiceID,
        "transaction_id": invoice.TransactionID,
        "invoice_html":   htmlContent,
        "print_url":      fmt.Sprintf("/invoice/print-html?transaction_id=%s&auto_print=%t", request.TransactionID, request.AutoPrint),
        "total_amount":   invoice.Total,
        "message":        "Invoice generated successfully",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Update invoice payment status
func UpdateInvoicePaymentStatus(w http.ResponseWriter, r *http.Request) {
    // Remove enableCors(&w) - handled by main.go now
    
    var request struct {
        TransactionID string `json:"transaction_id"`
        PaymentStatus string `json:"payment_status"`
        PaymentMethod string `json:"payment_method"`
        PaymentDate   string `json:"payment_date"`
        Notes         string `json:"notes"`
    }

    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil {
        http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
        return
    }

    response := map[string]interface{}{
        "message":        "Payment status updated successfully",
        "transaction_id": request.TransactionID,
        "payment_status": request.PaymentStatus,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Get invoice data (refactored from existing function)
// Update the SQL query in getInvoiceData to remove the kategori column

func getInvoiceData(transactionID string) (*Invoice, error) {
    fmt.Printf("Getting invoice data for transaction: %s\n", transactionID)
    
    db := Koneksi()
    if db == nil {
        return nil, fmt.Errorf("database connection failed")
    }
    defer db.Close()

    // Test database connection
    err := db.Ping()
    if err != nil {
        return nil, fmt.Errorf("database ping failed: %v", err)
    }

    // Updated SQL query - removed p.kategori and COALESCE for it
    sqlQuery := `
        SELECT 
            td.transaksi_id, 
            td.nama_produk, 
            td.harga AS harga_jual, 
            COALESCE(p.harga_beli, 0) as harga_beli, 
            td.jumlah_terjual, 
            td.total_harga, 
            th.tanggal_transaksi,
            th.total_item
        FROM transaksi_detail td
        INNER JOIN produk p ON td.nama_produk = p.nama
        INNER JOIN transaksi_header th ON td.transaksi_id = th.transaksi_id
        WHERE td.transaksi_id = $1
        ORDER BY td.nama_produk
    `

    fmt.Printf("Executing SQL query for transaction: %s\n", transactionID)
    rows, err := db.Query(sqlQuery, transactionID)
    if err != nil {
        return nil, fmt.Errorf("database query error: %v", err)
    }
    defer rows.Close()

    var invoice Invoice
    var items []InvoiceItem
    var subtotal float64
    invoiceCounter := 1

    // Business information
    invoice.BusinessInfo = BusinessInfo{
        Name:        "BINTANG JAYA FROZEN FOOD",
        Address:     "Jl. Raya Frozen No. 123, Jakarta Selatan 12345",
        Phone:       "+62 21 1234 5678",
        Email:       "info@bintangjaya.com",
        TaxID:       "01.234.567.8-901.000",
        Logo:        "/assets/logo.png",
        Website:     "www.bintangjaya.com",
        BankAccount: "BCA 1234567890 a.n. Bintang Jaya",
    }

    rowCount := 0
    for rows.Next() {
        rowCount++
        var transID string
        var productName string
        var hargaJual float64
        var hargaBeli float64
        var quantity int
        var totalPrice float64
        var transactionDate time.Time
        var totalItems int

        // Updated Scan call - removed category parameter
        err := rows.Scan(&transID, &productName, &hargaJual, &hargaBeli, &quantity, &totalPrice, &transactionDate, &totalItems)
        if err != nil {
            return nil, fmt.Errorf("row scan error: %v", err)
        }

        // Set invoice header info (only once)
        if invoice.TransactionID == "" {
            invoice.InvoiceID = fmt.Sprintf("INV-%s-%s", transID, transactionDate.Format("20060102"))
            invoice.TransactionID = transID
            invoice.InvoiceDate = transactionDate
            invoice.DueDate = transactionDate.AddDate(0, 0, 30)
            invoice.PaymentStatus = "PAID"
            invoice.PaymentMethod = "CASH"
            invoice.CashierName = "Kasir 1"
            invoice.InvoiceNumber = invoiceCounter
            invoice.CustomerName = "Walk-in Customer"
        }

        // Add item to invoice - use supplier name as category instead
        // Get the supplier name from produk table (visible in your database output)
        var category string
        if strings.Contains(productName, "Ayam") || strings.Contains(productName, "Sosis") {
            category = "Frozen Food"
        } else if strings.Contains(productName, "Flower") {
            category = "Seafood"
        } else {
            category = "General"
        }

        item := InvoiceItem{
            ProductName: productName,
            Quantity:    quantity,
            UnitPrice:   hargaJual,
            TotalPrice:  totalPrice,
            Description: fmt.Sprintf("Frozen food item - %s", productName),
            Category:    category,  // Using category based on product name
        }
        items = append(items, item)
        subtotal += totalPrice
    }

    fmt.Printf("Processed %d rows for transaction %s\n", rowCount, transactionID)

    if len(items) == 0 {
        return nil, fmt.Errorf("transaction not found: %s", transactionID)
    }

    invoice.Items = items
    invoice.Subtotal = subtotal
    invoice.Tax = subtotal * 0.11
    invoice.Discount = 0.0
    invoice.Total = invoice.Subtotal + invoice.Tax - invoice.Discount

    fmt.Printf("Invoice data prepared successfully: %d items, total: %.2f\n", len(items), invoice.Total)
    return &invoice, nil
}

// Fixed HTML template for POS paper printing
func generateInvoiceHTMLTemplate(invoice *Invoice) string {
    if invoice == nil {
        return "Error: Invoice data is nil"
    }

    fmt.Printf("Generating HTML template for invoice: %s\n", invoice.InvoiceID)

    // Updated HTML template with fixed formatting
    htmlTemplate := `
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice {{.InvoiceID}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Courier New', monospace;
            font-size: 12px;
            line-height: 1.4;
            color: #000;
            background: white;
            width: 80mm;
            margin: 0 auto;
        }
        
        .receipt-container {
            width: 100%;
            padding: 5mm;
            background: white;
        }
        
        .text-center {
            text-align: center;
        }
        
        .text-left {
            text-align: left;
        }
        
        .text-right {
            text-align: right;
        }
        
        .bold {
            font-weight: bold;
        }
        
        .header {
            text-align: center;
            margin-bottom: 10px;
            border-bottom: 1px dashed #000;
            padding-bottom: 10px;
        }
        
        .business-name {
            font-size: 16px;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .business-details {
            font-size: 10px;
            line-height: 1.2;
        }
        
        .invoice-info {
            margin: 10px 0;
            font-size: 11px;
        }
        
        .section-divider {
            border-bottom: 1px dashed #000;
            margin: 10px 0;
            padding-bottom: 5px;
        }
        
        .customer-info {
            margin: 10px 0;
            font-size: 11px;
        }
        
        .items-section {
            margin: 10px 0;
        }
        
        .item-row {
            margin-bottom: 8px;
            font-size: 11px;
        }
        
        .item-name {
            font-weight: bold;
            margin-bottom: 2px;
        }
        
        .item-details {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .qty-price {
            white-space: nowrap;
        }
        
        .totals-section {
            margin-top: 15px;
            border-top: 1px dashed #000;
            padding-top: 10px;
            font-size: 11px;
        }
        
        .total-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 3px;
        }
        
        .grand-total {
            font-weight: bold;
            font-size: 14px;
            border-top: 1px solid #000;
            padding-top: 5px;
            margin-top: 5px;
        }
        
        .footer {
            margin-top: 15px;
            text-align: center;
            font-size: 10px;
            border-top: 1px dashed #000;
            padding-top: 10px;
        }
        
        .thank-you {
            margin: 10px 0;
            font-weight: bold;
        }
        
        .contact-info {
            margin-top: 10px;
            line-height: 1.3;
        }
        
        /* Print specific styles for thermal paper */
        @media print {
            body {
                background: white;
                -webkit-print-color-adjust: exact;
                print-color-adjust: exact;
                margin: 0;
                padding: 0;
            }
            
            .receipt-container {
                margin: 0;
                padding: 2mm;
            }
            
            @page {
                margin: 0;
                size: 80mm auto;
            }
        }
        
        /* Additional styles for better readability */
        .dotted-line {
            border-bottom: 1px dotted #000;
            margin: 5px 0;
        }
        
        .payment-status {
            background: #000;
            color: white;
            padding: 2px 5px;
            font-size: 10px;
            display: inline-block;
            margin: 5px 0;
        }
    </style>
</head>
<body>
    <div class="receipt-container">
        <!-- Header -->
        <div class="header">
            <div class="business-name">{{.BusinessInfo.Name}}</div>
            <div class="business-details">
                {{.BusinessInfo.Address}}<br>
                Tel: {{.BusinessInfo.Phone}}<br>
                Email: {{.BusinessInfo.Email}}<br>
                NPWP: {{.BusinessInfo.TaxID}}
            </div>
        </div>

        <!-- Invoice Information -->
        <div class="invoice-info">
            <div><strong>INVOICE: {{.InvoiceID}}</strong></div>
            <div>Transaksi ID: {{.TransactionID}}</div>
            <div>Tanggal: {{.InvoiceDate.Format "02/01/2006 15:04"}}</div>
            <div>Kasir: {{.CashierName}}</div>
        </div>

        <!-- Customer Information -->
        <div class="customer-info section-divider">
            <div><strong>Pelanggan: {{.CustomerName}}</strong></div>
            {{if .CustomerPhone}}<div>Tel: {{.CustomerPhone}}</div>{{end}}
            <div>Pembayaran: {{.PaymentMethod}}</div>
            <div class="payment-status">{{.PaymentStatus}}</div>
        </div>

        <!-- Items -->
        <div class="items-section">
            {{range $index, $item := .Items}}
            <div class="item-row">
                <div class="item-name">{{$item.ProductName}}</div>
                <div class="item-details">
                    <span>{{$item.Quantity}} x {{formatCurrency $item.UnitPrice}}</span>
                    <span class="bold">{{formatCurrency $item.TotalPrice}}</span>
                </div>
                {{if $item.Category}}<div style="font-size: 9px; color: #666;">[{{$item.Category}}]</div>{{end}}
            </div>
            {{end}}
        </div>

        <!-- Totals -->
        <div class="totals-section">
            <div class="total-row">
                <span>Subtotal:</span>
                <span>{{formatCurrency .Subtotal}}</span>
            </div>
            <div class="total-row">
                <span>PPN (11%):</span>
                <span>{{formatCurrency .Tax}}</span>
            </div>
            {{if ne .Discount 0.0}}
            <div class="total-row">
                <span>Diskon:</span>
                <span>-{{formatCurrency .Discount}}</span>
            </div>
            {{end}}
            <div class="total-row grand-total">
                <span>TOTAL:</span>
                <span>{{formatCurrency .Total}}</span>
            </div>
        </div>

        <!-- Footer -->
        <div class="footer">
            <div class="thank-you">
                *** TERIMA KASIH ***<br>
                Selamat Berbelanja Kembali
            </div>
            
            <div class="dotted-line"></div>
            
            <div class="contact-info">
                <strong>Customer Service:</strong><br>
                WhatsApp: +62 812-3456-7890<br>
                Instagram: @bintangjaya_frozen<br>
                <br>
                <strong>Jam Operasional:</strong><br>
                Sen-Sab: 08:00-22:00<br>
                Minggu: 09:00-21:00<br>
                <br>
                <div style="font-size: 9px;">
                    Barang yang sudah dibeli<br>
                    tidak dapat dikembalikan<br>
                    Simpan struk sebagai bukti
                </div>
            </div>
        </div>
        
        <div class="dotted-line"></div>
        <div class="text-center" style="font-size: 9px; margin-top: 10px;">
            Powered by Bintang Jaya POS
        </div>
    </div>

    <script>
        // Auto print when requested
        if (window.location.search.includes('auto_print=true')) {
            window.onload = function() {
                setTimeout(function() {
                    window.print();
                }, 500);
            };
        }
    </script>
</body>
</html>`

    // Parse and execute template with proper error handling and FIXED formatCurrency function
    tmpl, err := template.New("invoice").Funcs(template.FuncMap{
        "formatCurrency": func(amount float64) string {
            // Fix the currency formatting to avoid showing float64 representation
            return fmt.Sprintf("Rp %s", formatNumber(amount))
        },
        "add": func(a, b int) int {
            return a + b
        },
        "ne": func(a, b float64) bool {
            return a != b
        },
    }).Parse(htmlTemplate)

    if err != nil {
        fmt.Printf("Template parsing error: %v\n", err)
        return fmt.Sprintf("Error parsing template: %v", err)
    }

    var buf bytes.Buffer
    err = tmpl.Execute(&buf, invoice)
    if err != nil {
        fmt.Printf("Template execution error: %v\n", err)
        return fmt.Sprintf("Error executing template: %v", err)
    }

    htmlContent := buf.String()
    fmt.Printf("HTML template generated successfully, length: %d\n", len(htmlContent))
    return htmlContent
}

// Add a helper function to format numbers with thousand separators
func formatNumber(num float64) string {
    // Convert to integer if it's a whole number
    if num == float64(int(num)) {
        return addThousandSeparator(int(num))
    }
    
    // Otherwise format with 0 decimal places and convert to int
    return addThousandSeparator(int(num))
}

// Helper function for thousand separator
func addThousandSeparator(n int) string {
    in := strconv.FormatInt(int64(n), 10)
    out := make([]byte, 0, len(in)+(len(in)-2+int(in[0]/'0'))/3)
    
    if in[0] == '-' {
        out = append(out, '-')
        in = in[1:]
    }

    // Add digits with thousand separators
    for i := 0; i < len(in); i++ {
        if i > 0 && (len(in)-i)%3 == 0 {
            out = append(out, '.')
        }
        out = append(out, in[i])
    }
    
    return string(out)
}

// Add the missing getIntOrDefault function and fix unused variable

// Add this function at the end of your file
func getIntOrDefault(str string, defaultVal int) int {
    if str == "" {
        return defaultVal
    }
    if val, err := strconv.Atoi(str); err == nil {
        return val
    }
    return defaultVal
}