package rest

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/ixofoundation/ixo-blockchain/x/did/exported"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"

	"github.com/ixofoundation/ixo-blockchain/x/did/internal/keeper"
	"github.com/ixofoundation/ixo-blockchain/x/did/internal/types"
	"github.com/ixofoundation/ixo-blockchain/x/ixo"

	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/did", createDidRequestHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/credential", addCredentialRequestHandler(cliCtx)).Methods("POST")
}

type MsgAddDid struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	Did     exported.Did `json:"did" yaml:"did"`
	PubKey  string       `json:"pubKey" yaml:"pubKey"`
}

func createDidRequestHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req MsgAddDid

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		msg := types.NewMsgAddDid(req.Did, req.PubKey)
		//this is  referred  from cosmos version v0.37.0
		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func addCredentialRequestHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		did := r.URL.Query().Get("did")
		didDocParam := r.URL.Query().Get("signerDidDoc")
		mode := r.URL.Query().Get("mode")
		cliCtx = cliCtx.WithBroadcastMode(mode)

		_, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute,
			keeper.QueryDidDoc, did), nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("The did is not found"))
			return
		}

		ixoDid, err := types.UnmarshalIxoDid(didDocParam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		t := time.Now()
		issued := t.Format(time.RFC3339)

		credTypes := []string{"Credential", "ProofOfKYC"}

		msg := types.NewMsgAddCredential(did, credTypes, ixoDid.Did, issued)

		output, err := ixo.CompleteAndBroadcastTxRest(cliCtx, msg, ixoDid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		rest.PostProcessResponse(w, cliCtx, output)
	}
}
