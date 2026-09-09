package budget

import "fmt"

type Envelope struct {
	Version      int                `json:"version"`
	Scope        string             `json:"scope"`
	Limits       map[string]float64 `json:"limits"`
	Usage        map[string]float64 `json:"usage"`
	Reservations []Reservation      `json:"reservations"`
	Mode         string             `json:"mode"`
}
type Reservation struct {
	ID       string             `json:"id"`
	Scope    string             `json:"scope"`
	Amount   map[string]float64 `json:"amount"`
	Released bool               `json:"released"`
}

func New(scope string, limits map[string]float64) Envelope {
	return Envelope{Version: 1, Scope: scope, Limits: limits, Usage: map[string]float64{}, Reservations: []Reservation{}, Mode: "soft"}
}
func (e Envelope) Reserve(id, scope string, amount map[string]float64) (Envelope, error) {
	for key, value := range amount {
		if e.Limits[key] > 0 && e.Usage[key]+value > e.Limits[key] {
			return e, fmt.Errorf("budget limit exceeded for %s", key)
		}
	}
	e.Reservations = append(e.Reservations, Reservation{ID: id, Scope: scope, Amount: amount})
	return e, nil
}
func (e Envelope) Consume(amount map[string]float64) (Envelope, error) {
	for key, value := range amount {
		if e.Limits[key] > 0 && e.Usage[key]+value > e.Limits[key] {
			if e.Mode == "hard" {
				return e, fmt.Errorf("hard budget exhausted for %s", key)
			}
		}
		e.Usage[key] += value
	}
	return e, nil
}
func (e Envelope) Remaining(key string) float64 { return e.Limits[key] - e.Usage[key] }
