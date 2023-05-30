package nftuser

type userWithdrawRequest struct {
	Amount string `json:"amount"`
}

type userWithdrawResponse struct {
	TxID  string `json:"tx_id,omitempty"`
	Error string `json:"error,omitempty"`
}

func userWithdraw(userid string, req *userWithdrawRequest, res *userWithdrawResponse) error {
	return nil
}
