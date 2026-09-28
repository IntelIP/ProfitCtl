package cmd

import (
 "bytes"
 "context"
 "encoding/json"
 "math"
 "os"
 "path/filepath"
 "testing"

 costv1 "github.com/IntelIP/ProfitCtl/pkg/contracts/cost/v1"
 variancev1 "github.com/IntelIP/ProfitCtl/pkg/variance/v1"
 "github.com/spf13/cobra"
 "github.com/stretchr/testify/require"
)

func varianceFixture(t *testing.T, id string, quantity, price, total float64, kind costv1.DriverKind) (costv1.CostObservation, costv1.CostDriver) {
 t.Helper()
 source := costv1.SourceReference{Type: costv1.SourceSyntheticFixture, ArtifactIdentity: "fixture://variance", CapturedAt: "2026-08-01T00:00:00Z"}
 e := costv1.Evidence{Kind: costv1.EvidencePredicted, Measurement: costv1.MeasurementSynthetic, Source: source, Confidence: costv1.ConfidenceLow, ConfidenceRationale: "deterministic fixture"}
 window := costv1.TimeWindow{Start:"2026-07-01T00:00:00Z", End:"2026-08-01T00:00:00Z"}
 q := costv1.Quantity{Value:quantity, Unit:"command"}
 p := costv1.UnitPrice{Amount:costv1.Money{Amount:price, Currency:"USD"}, Per:costv1.Quantity{Value:1, Unit:"command"}}
 o := costv1.CostObservation{SchemaVersion:costv1.SchemaVersion, ID:id, DriverIDs:[]string{"driver"}, Window:window, Quantity:q, UnitPrice:p, TotalCost:costv1.Money{Amount:total, Currency:"USD"}, Dimensions:costv1.Dimensions{Workload:"worker"}, Evidence:costv1.ClaimEvidence{Quantity:e, UnitPrice:e, TotalCost:e}}
 d := costv1.CostDriver{SchemaVersion:costv1.SchemaVersion, ID:"driver", Name:"fixture driver", Kind:kind, Quantity:q, UnitPrice:p, Window:window, Dimensions:o.Dimensions, Evidence:costv1.DriverEvidence{Quantity:e, UnitPrice:e}}
 if kind == costv1.DriverVariable || kind == costv1.DriverCadence { d.Per = &costv1.Quantity{Value:1, Unit:"second"} }
 require.NoError(t, o.Validate())
 require.NoError(t, d.Validate())
 return o, d
}

func varianceWriteJSON(t *testing.T, file string, value any) {
 t.Helper()
 data, err := json.Marshal(value)
 require.NoError(t, err)
 require.NoError(t, os.WriteFile(file, data, 0600))
}
func varianceExecute(t *testing.T, command *cobra.Command, args ...string) ([]byte, error) {
 t.Helper()
 var output bytes.Buffer
 command.SetOut(&output)
 command.SetErr(&bytes.Buffer{})
 command.SetArgs(args)
 err := command.Execute()
 return output.Bytes(), err
}

func TestCostDiffExplainJSON(t *testing.T) {
 for _, kind := range []costv1.DriverKind{costv1.DriverFixed, costv1.DriverVariable, costv1.DriverCadence, costv1.DriverConcurrency} {
  t.Run(string(kind),func(t *testing.T){
   f, fd := varianceFixture(t,"forecast",10,2,20,kind)
   a, ad := varianceFixture(t,"actual",20,2,40,kind)
   dir := t.TempDir()
   input, diffPath := filepath.Join(dir,"input.json"),filepath.Join(dir,"diff.json")
   varianceWriteJSON(t,input,variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}}})
   payload, err := varianceExecute(t,newCostDiffCommand(),"--input",input,"--output",diffPath)
   require.NoError(t,err)
   disk, err := os.ReadFile(diffPath); require.NoError(t,err); require.Equal(t,payload,disk)
   var diff variancev1.Result
   require.NoError(t,json.Unmarshal(payload,&diff))
   require.Equal(t,variancev1.SchemaVersion,diff.SchemaVersion)
   require.Equal(t,f.Window,diff.Window)
   require.Equal(t,20.0,diff.AbsoluteVariance.Amount)
   require.InDelta(t,0,diff.Residual.Amount,variancev1.Rounding)
   require.Equal(t,"USD",diff.Contributions[0].Amount.Currency)
   require.NotEmpty(t,diff.Contributions[0].Sources[0].ArtifactIdentity)
   require.NotEmpty(t,diff.Contributions[0].Sources[0].CapturedAt)
   explain,err := varianceExecute(t,newCostExplainCommand(),"--input",diffPath)
   require.NoError(t,err)
   var report ExplainReport
   require.NoError(t,json.Unmarshal(explain,&report))
   require.Len(t,report.DriverContributions,1)
   require.NotEmpty(t,report.DriverContributions[0].RemedialActions)
   require.Greater(t,report.DriverContributions[0].ConfidenceScore,0.0)
   if kind == costv1.DriverCadence || kind == costv1.DriverConcurrency {
    require.Equal(t,"unknown",report.DriverContributions[0].Attribution)
    require.Contains(t,report.DriverContributions[0].MissingEvidence,"exact delivery receipt")
   }
  })
 }
}

func TestCostVariancePriceAndReceipts(t *testing.T) {
 f,fd := varianceFixture(t,"forecast",10,2,20,costv1.DriverVariable)
 a,ad := varianceFixture(t,"actual",10,3,30,costv1.DriverVariable)
 result,err := variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}}})
 require.NoError(t,err)
 require.Greater(t,math.Abs(result.Contributions[0].UnitPrice),math.Abs(result.Contributions[0].Quantity))
 require.InDelta(t,0,result.Residual.Amount,variancev1.Rounding)
 f,fd = varianceFixture(t,"forecast",10,2,20,costv1.DriverCadence)
 a,ad = varianceFixture(t,"actual",20,2,40,costv1.DriverCadence)
 receipt := variancev1.Receipt{DriverID:"driver",Kind:"deployment",ForecastObservationID:f.ID,ActualObservationID:a.ID,Source:costv1.SourceReference{Type:costv1.SourceRuntimeLedger,ArtifactIdentity:"deployment://exact",CapturedAt:"2026-08-01T00:00:00Z"}}
 result,err = variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}},Receipts:[]variancev1.Receipt{receipt}})
 require.NoError(t,err)
 require.Equal(t,"modeled",result.Contributions[0].Attribution)
 require.Len(t,result.Contributions[0].Receipts,1)
 receipt.ActualObservationID = "other"
 _,err = variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}},Receipts:[]variancev1.Receipt{receipt}})
 require.Error(t,err)
}

func TestCostVarianceResidualZeroAndUnavailable(t *testing.T) {
 f,fd := varianceFixture(t,"forecast",10,2,21,costv1.DriverVariable)
 a,ad := varianceFixture(t,"actual",20,3,65,costv1.DriverVariable)
 result,err := variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}}})
 require.NoError(t,err)
 require.InDelta(t,4,result.Residual.Amount,variancev1.Rounding)
 report,err := Explain(context.Background(),&result)
 require.NoError(t,err)
 require.Equal(t,4.0,report.Residual.Amount)
 require.NotEmpty(t,report.MissingEvidence)
 result.AbsoluteVariance.Amount++
 _,err = Explain(context.Background(),&result)
 require.ErrorContains(t,err,"reconcile")
 f.TotalCost.Amount = 0
 result,err = variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a})
 require.NoError(t,err)
 require.Nil(t,result.PercentageVariance)
 unavailable,err := Diff(context.Background(),&f,nil)
 require.NoError(t,err)
 require.Equal(t,"unavailable",unavailable.Status)
 require.Nil(t,unavailable.Actual)
 require.Nil(t,unavailable.AbsoluteVariance)
 _,err = Explain(context.Background(),unavailable)
 require.NoError(t,err)
}

func TestCostDiffInvalidInputAndAliases(t *testing.T) {
 f,_ := varianceFixture(t,"forecast",10,2,20,costv1.DriverFixed)
 a,_ := varianceFixture(t,"actual",10,3,30,costv1.DriverFixed)
 dir:=t.TempDir(); forecast:=filepath.Join(dir,"forecast.json"); actual:=filepath.Join(dir,"actual.json")
 varianceWriteJSON(t,forecast,f); varianceWriteJSON(t,actual,a)
 _,err:=varianceExecute(t,newCostDiffCommand(),"--forecast",forecast,"--actual",actual)
 require.NoError(t,err)
 _,err=varianceExecute(t,newCostDiffCommand(),"--forecast",forecast,"--actual",actual,"--output",forecast)
 require.Equal(t,2,ExitCode(err))
 _,err=varianceExecute(t,newCostDiffCommand())
 require.Equal(t,2,ExitCode(err))
 require.NoError(t,os.WriteFile(actual,[]byte("{}{}"),0600))
 _,err=varianceExecute(t,newCostDiffCommand(),"--forecast",forecast,"--actual",actual)
 require.Equal(t,2,ExitCode(err))
 a.Evidence.Quantity = costv1.Evidence{}
 varianceWriteJSON(t,actual,a)
 _,err=varianceExecute(t,newCostDiffCommand(),"--forecast",forecast,"--actual",actual)
 require.Equal(t,2,ExitCode(err))
 _,err=varianceExecute(t,newCostExplainCommand(),"--input",actual)
 require.Equal(t,2,ExitCode(err))
 _,err=varianceExecute(t,newCostExplainCommand(),"--input",actual,"--output",actual)
 require.Equal(t,2,ExitCode(err))
}

func TestCostDiffUnavailableAndZeroJSON(t *testing.T) {
 f,_ := varianceFixture(t,"forecast",0,2,0,costv1.DriverFixed)
 a,_ := varianceFixture(t,"actual",1,2,2,costv1.DriverFixed)
 dir:=t.TempDir(); input:=filepath.Join(dir,"input.json")
 varianceWriteJSON(t,input,variancev1.Input{Forecast:f,Actual:&a})
 payload,err:=varianceExecute(t,newCostDiffCommand(),"--input",input)
 require.NoError(t,err)
 var result variancev1.Result
 require.NoError(t,json.Unmarshal(payload,&result))
 require.Nil(t,result.PercentageVariance)
 varianceWriteJSON(t,input,variancev1.Input{Forecast:f,UnavailableReason:"invoice not delivered"})
 payload,err=varianceExecute(t,newCostDiffCommand(),"--input",input)
 require.NoError(t,err)
 require.NoError(t,json.Unmarshal(payload,&result))
 require.Equal(t,"unavailable",result.Status)
 require.Nil(t,result.Actual)
 require.Nil(t,result.Residual)
}

func TestCostExplainTotalCurrencies(t *testing.T) {
 f,_ := varianceFixture(t,"forecast",10,2,20,costv1.DriverFixed)
 a,_ := varianceFixture(t,"actual",10,3,30,costv1.DriverFixed)
 base,err := variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a})
 require.NoError(t,err)
 for _, field := range []string{"actual", "absolute_variance", "residual"} {
  t.Run(field,func(t *testing.T){
   result := base
   actual, absolute, residual := *base.Actual, *base.AbsoluteVariance, *base.Residual
   result.Actual, result.AbsoluteVariance, result.Residual = &actual, &absolute, &residual
   switch field {
   case "actual": result.Actual.Currency = "EUR"
   case "absolute_variance": result.AbsoluteVariance.Currency = "EUR"
   case "residual": result.Residual.Currency = "EUR"
   }
   _,err := Explain(context.Background(),&result)
   require.ErrorContains(t,err,field)
   require.ErrorContains(t,err,"EUR")
   require.ErrorContains(t,err,"USD")
   require.NotContains(t,err.Error(),"reconcile")
  })
 }
 report,err := Explain(context.Background(),&base)
 require.NoError(t,err)
 require.Equal(t,"USD",report.Actual.Currency)
 require.Equal(t,"USD",report.VarianceAbsolute.Currency)
 require.Equal(t,"USD",report.Residual.Currency)
}

func TestCostExplainCapsStrongDriversByTotalEvidence(t *testing.T) {
 f,fd := varianceFixture(t,"forecast",10,2,20,costv1.DriverVariable)
 a,ad := varianceFixture(t,"actual",20,2,40,costv1.DriverVariable)
 strong := costv1.Evidence{Kind:costv1.EvidenceObserved,Measurement:costv1.MeasurementMeasured,Source:costv1.SourceReference{Type:costv1.SourceTelemetry,ArtifactIdentity:"telemetry://driver",CapturedAt:"2026-08-01T00:00:00Z"},Confidence:costv1.ConfidenceHigh,ConfidenceRationale:"measured driver"}
 fd.Evidence.Quantity,ad.Evidence.Quantity = strong,strong
 strong.Source.Type = costv1.SourceInvoice
 strong.Source.ArtifactIdentity = "invoice://driver"
 strong.Kind = costv1.EvidenceBilled
 fd.Evidence.UnitPrice,ad.Evidence.UnitPrice = strong,strong
 result,err := variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}}})
 require.NoError(t,err)
 require.Equal(t,costv1.ConfidenceHigh,result.Contributions[0].Confidence)
 report,err := Explain(context.Background(),&result)
 require.NoError(t,err)
 require.Equal(t,costv1.ConfidenceLow,report.Confidence)
 require.LessOrEqual(t,report.ConfidenceScore,.35)
 require.Equal(t,costv1.ConfidenceHigh,report.DriverContributions[0].Confidence)
 require.Contains(t,report.MissingEvidence,"forecast total cost: high-confidence claim evidence unavailable")
 require.Contains(t,report.MissingEvidence,"actual total cost: high-confidence claim evidence unavailable")
}

func TestCostExplainSourceLabelsCannotPromoteTotalConfidence(t *testing.T) {
 for _, sourceType := range []costv1.SourceType{costv1.SourceInvoice, costv1.SourceProfitCtlDerived} {
  t.Run(string(sourceType),func(t *testing.T){
   f,fd := varianceFixture(t,"forecast",10,2,20,costv1.DriverVariable)
   a,ad := varianceFixture(t,"actual",20,2,40,costv1.DriverVariable)
   for _, observation := range []*costv1.CostObservation{&f,&a} {
    observation.Evidence.TotalCost.Source.Type = sourceType
    if sourceType == costv1.SourceInvoice {
     observation.Evidence.TotalCost.Kind = costv1.EvidenceBilled
     observation.Evidence.TotalCost.Measurement = costv1.MeasurementMeasured
    } else {
     observation.Evidence.TotalCost.Measurement = costv1.MeasurementDerived
    }
   }
   strong := costv1.Evidence{Kind:costv1.EvidenceObserved,Measurement:costv1.MeasurementMeasured,Source:costv1.SourceReference{Type:costv1.SourceTelemetry,ArtifactIdentity:"telemetry://driver",CapturedAt:"2026-08-01T00:00:00Z"},Confidence:costv1.ConfidenceHigh,ConfidenceRationale:"measured driver"}
   fd.Evidence.Quantity,ad.Evidence.Quantity = strong,strong
   strong.Source.Type = costv1.SourceInvoice
   strong.Source.ArtifactIdentity = "invoice://driver"
   strong.Kind = costv1.EvidenceBilled
   fd.Evidence.UnitPrice,ad.Evidence.UnitPrice = strong,strong
   result,err := variancev1.Compare(variancev1.Input{Forecast:f,Actual:&a,Drivers:[]variancev1.DriverPair{{Forecast:fd,Actual:ad}}})
   require.NoError(t,err)
   require.Equal(t,"available",result.Status)
   require.InDelta(t,0,result.Residual.Amount,variancev1.Rounding)
   require.Equal(t,costv1.ConfidenceHigh,result.Contributions[0].Confidence)
   require.Greater(t,result.Contributions[0].ConfidenceScore,.7)
   require.Equal(t,sourceType,result.Sources[0].Type)
   require.Equal(t,sourceType,result.Sources[1].Type)
   report,err := Explain(context.Background(),&result)
   require.NoError(t,err)
   require.Equal(t,costv1.ConfidenceLow,report.Confidence)
   require.LessOrEqual(t,report.ConfidenceScore,.35)
   require.Equal(t,result.Contributions[0].ConfidenceScore,report.DriverContributions[0].ConfidenceScore)
   require.Contains(t,report.MissingEvidence,"forecast total cost: high-confidence claim evidence unavailable")
   require.Contains(t,report.MissingEvidence,"actual total cost: high-confidence claim evidence unavailable")
  })
 }
}

func TestCostExplainInvalidReconciliation(t *testing.T) {
 result:=variancev1.Result{SchemaVersion:variancev1.SchemaVersion,Status:"available",Actual:&costv1.Money{Amount:2,Currency:"USD"},AbsoluteVariance:&costv1.Money{Amount:2,Currency:"USD"},Residual:&costv1.Money{Amount:math.NaN(),Currency:"USD"}}
 _,err:=Explain(context.Background(),&result)
 require.Error(t,err)
}
