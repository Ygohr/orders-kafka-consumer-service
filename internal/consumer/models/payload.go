package models

type OrderPayload struct {
	OrdemDeVenda string `json:"ordemDeVenda"`
	EtapaAtual   string `json:"etapaAtual"`
}
