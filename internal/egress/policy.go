package egress

type Class string

const (
	Public       Class = "public"
	Internal     Class = "internal"
	Confidential Class = "confidential"
	Restricted   Class = "restricted"
)

type Decision struct {
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason"`
	DataClass   Class  `json:"data_class"`
	Destination string `json:"destination"`
}

func Allow(data Class, destination string, local bool) Decision {
	if data == Restricted && !local {
		return Decision{Reason: "restricted data requires local destination", DataClass: data, Destination: destination}
	}
	return Decision{Allowed: true, Reason: "egress permitted by data policy", DataClass: data, Destination: destination}
}
