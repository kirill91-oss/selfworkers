package main

import (
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strconv"
)

const (
    rateIndividuals      = 0.04
    rateBusinesses       = 0.06
    reliefCap            = 10000.0
    reliefIndividualsPct = 0.01
    reliefBusinessPct    = 0.02
)

type CalculationResult struct {
    IncomeIndividuals float64
    IncomeBusinesses  float64
    TaxIndividuals    float64
    TaxBusinesses     float64
    ReliefApplied     float64
    RemainingRelief   float64
    TotalTax          float64
    HasResult         bool
    Error             string
}

var tmpl = template.Must(template.ParseFiles("templates/index.gohtml"))

func main() {
    http.HandleFunc("/", handleForm)
    log.Println("Server is running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleForm(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        renderTemplate(w, CalculationResult{})
    case http.MethodPost:
        if err := r.ParseForm(); err != nil {
            renderTemplate(w, CalculationResult{Error: "Не удалось обработать форму"})
            return
        }

        indIncome, err := parseAmount(r.FormValue("individual_income"))
        if err != nil {
            renderTemplate(w, CalculationResult{Error: "Некорректный доход от физических лиц"})
            return
        }

        bizIncome, err := parseAmount(r.FormValue("business_income"))
        if err != nil {
            renderTemplate(w, CalculationResult{Error: "Некорректный доход от юрлиц и ИП"})
            return
        }

        result := calculateTax(indIncome, bizIncome)
        renderTemplate(w, result)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func renderTemplate(w http.ResponseWriter, data CalculationResult) {
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
    }
}

func parseAmount(value string) (float64, error) {
    if value == "" {
        return 0, nil
    }
    v, err := strconv.ParseFloat(value, 64)
    if err != nil {
        return 0, err
    }
    if v < 0 {
        return 0, fmt.Errorf("negative amount")
    }
    return v, nil
}

func calculateTax(indIncome, bizIncome float64) CalculationResult {
    taxInd := indIncome * rateIndividuals
    taxBiz := bizIncome * rateBusinesses

    potentialRelief := indIncome*reliefIndividualsPct + bizIncome*reliefBusinessPct
    reliefApplied := mathMin(reliefCap, potentialRelief)
    totalTax := taxInd + taxBiz - reliefApplied
    if totalTax < 0 {
        totalTax = 0
    }

    return CalculationResult{
        IncomeIndividuals: indIncome,
        IncomeBusinesses:  bizIncome,
        TaxIndividuals:    taxInd,
        TaxBusinesses:     taxBiz,
        ReliefApplied:     reliefApplied,
        RemainingRelief:   reliefCap - reliefApplied,
        TotalTax:          totalTax,
        HasResult:         indIncome > 0 || bizIncome > 0,
    }
}

func mathMin(a, b float64) float64 {
    if a < b {
        return a
    }
    return b
}
