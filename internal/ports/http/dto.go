package http

type ResponseCoin struct {
	AggregationType string  `json:"aggregation_type,omitempty"`
	Title           string  `json:"title"`
	Cost            float64 `json:"cost"`
}

type Response struct {
	Coins []*ResponseCoin `json:"coins"`
}

type DtoErrResponse struct {
	Code    int    `json:"code"`
	Message string `json:"massage"`
}
