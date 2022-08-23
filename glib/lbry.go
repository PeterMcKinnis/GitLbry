package glib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

const lbryrpcServer = "http://localhost:5279"

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResult[T any] struct {
	Jsonrpc string    `json:"jsonrpc"`
	Result  *T        `json:"result"`
	Error   *rpcError `json:"error"`
	Id      int       `json:"id"`
}

type rpcRequest[T any] struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
	Id      int    `json:"id"`
}

type sdkPage[T any] struct {
	// Page number of the current items.
	Page int `json:"page"`

	// Number of items to show on a page.
	PageSize int `json:"page_size"`

	// Total number of pages.
	TotalItems int `json:"total_items"`

	// Total number of items
	TotalPages int `json:"total_pages"`

	Items []T `json:"items"`
}

type withError struct {
	Error *json.RawMessage `json:"error"`
}

// If the claim could not be loaded due to an error, returns a string
// with the error contents and true.  otherwise returns an empty string and
// false.
func (c *withError) GetError() error {

	if c.Error == nil {
		return nil
	}

	return errors.New("sdkError: " + string(*c.Error))

}

type sdkClaim struct {
	withError

	//Address        string `json:"address"`
	//Amount         string `json:"amount"`
	//CanonicalUrl   string `json:"canonical_url"`
	ClaimId string `json:"claim_id"`
	//Height         int    `json:"height"`
	//Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
	//Nout           int    `json:"nout"`
	//PermanentUrl   string `json:"permanent_url"`
	//ShortUrl       string `json:"short_url"`
	//Timestamp      int    `json:"timestamp"`
	//Txid           string `json:"txid"`
	//Type           string `json:"type"`
	//Value interface{}
	// determines the type of the 'value' field: 'channel', 'stream', etc"
	//ValueType string `json:"value_type"`

	SigningChannel *sdkClaim `json:"signing_channel"`

	// True, false or null for unknown
	IsMyOutput *bool `json:"is_my_output"`
}

func lbryResolve(url string) (*sdkClaim, error) {

	type arg struct {
		// This is plural the server accepts both a single string or a list of strings
		Urls              string `json:"urls"`
		IncludeIsMyOutput bool   `json:"include_is_my_output"`
	}

	result, err := rpcCall[arg, map[string]sdkClaim]("resolve", arg{
		Urls:              url,
		IncludeIsMyOutput: true,
	})

	if err != nil {
		return nil, err
	}

	claim, ok := result[url]
	if !ok {
		return nil,
			errors.New("error parsing output of lbryResolve, unexpected format")
	}

	err = claim.GetError()
	if err != nil {
		return nil, err
	}

	return &claim, nil
}

func lbryGet(uri string, fileName string) error {

	type arg struct {
		Uri      string `json:"uri"`
	}

	type out struct {
		withError
		BlobsRemaining int    `json:"blobs_remaining"`
		DownloadPath string `json:"download_path"`
	}

	o, err := rpcCall[arg, out]("get", arg{
		Uri:      uri,
	})

	if err != nil {
		return err
	}

	err = o.GetError()
	if err != nil {
		return err
	}

	// This happens sometimes ... should I wait here??
	// not really sure
	if o.BlobsRemaining > 0 {
		return errors.New("error reading document. It may been changed recently.  Try again in about 1 minute. %v")
	}

	// Move file to where we want it.  Note we can ask the lbry server do to this but it uses
	// a different root path.  Easier to handle this way
	err = os.Rename(o.DownloadPath, fileName);
	if err != nil {
		return err;
	}
	
	return nil

}

type lbryChannel struct {
	name string
	id   string
}

func lbryStreamUpdate(claimId string, filePath string) error {

	type arg struct {
		ClaimId  string `json:"claim_id"`
		FilePath string `json:"file_path"`
		Blocking bool   `json:"blocking"`
	}

	type out struct {
		withError
		// Not interested in any of the outputs
	}

	o, err := rpcCall[arg, out]("stream_update", arg{
		FilePath: filePath,
		ClaimId:  claimId,
		Blocking: true,
	})

	err = o.GetError()
	if err != nil {
		return err
	}

	return err
}

func lbryStreamCreateOnChannel(name string, channelId string, bid string, filePath string) error {

	type arg struct {
		Name      string `json:"name"`
		Bid       string `json:"bid"`
		FilePath  string `json:"file_path"`
		ChannelId string `json:"channel_id"`
		Blocking  bool   `json:"blocking"`
	}

	type out struct {
		withError
		// Not interested in any of the outputs
	}

	o, err := rpcCall[arg, out]("stream_create", arg{
		Name:      name,
		Bid:       bid,
		FilePath:  filePath,
		ChannelId: channelId,
		Blocking:  true,
	})

	err = o.GetError()
	if err != nil {
		return err
	}

	return err
}

func lbryStreamCreateForBundle(name string, channelId string, description string, bid string, filePath string) error {

	type arg struct {
		Name        string `json:"name"`
		Bid         string `json:"bid"`
		FilePath    string `json:"file_path"`
		ChannelId   string `json:"channel_id"`
		Description string `json:"description"`
		Blocking    bool   `json:"blocking"`
	}

	type out struct {
		withError
		// Not interested in any of the outputs other than error
	}

	o, err := rpcCall[arg, out]("stream_create", arg{
		Name:      name,
		Bid:       bid,
		FilePath:  filePath,
		ChannelId: channelId,
		Blocking:  true,
	})

	err = o.GetError()
	if err != nil {
		return err
	}

	return err
}

func lbryStreamCreate(name string, bid string, filePath string) error {

	type arg struct {
		Name     string `json:"name"`
		Bid      string `json:"bid"`
		FilePath string `json:"file_path"`
		Blocking bool   `json:"blocking"`
	}

	type out struct {
		withError
		// Not interested in any of the outputs
	}

	o, err := rpcCall[arg, out]("stream_create", arg{
		Name:     name,
		Bid:      bid,
		FilePath: filePath,
		Blocking: true,
	})

	err = o.GetError()
	if err != nil {
		return err
	}

	return err
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func rpcCall[Req any, Res any](method string, req Req) (Res, error) {

	r := rpcRequest[Req]{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  req,
		Id:      0,
	}

	rJson, err := json.MarshalIndent(&r, "", "  ")
	if err != nil {
		return zero[Res](), errors.Wrap(err, "error marshaling request to json")
	}

	OutPrintf("rcp call: %v\n%v\n", method, string(rJson))

	resp, err := http.Post(lbryrpcServer, "application/json", bytes.NewReader(rJson))
	if err != nil {
		return zero[Res](), errors.Wrap(err, "error durring http.POST to lbrynet rpc server")
	}

	buf := streamToByte(resp.Body)
	OutPrintf("result:\n%v\n", string(buf))

	var result rpcResult[Res]
	err = json.Unmarshal(buf, &result)
	if err != nil {
		return zero[Res](), errors.Wrap(err, "error decoding response from rpc server")
	}

	buf2, err := json.MarshalIndent(&result, "", "  ")
	debugPrintf(string(buf2))

	// Check that the server didn't return an error
	if result.Error != nil {
		return zero[Res](), errors.Errorf("lbry rpc server returned an error: %v %v ", result.Error.Code, result.Error.Message)
	}

	if result.Result == nil {
		return zero[Res](), errors.Errorf("lbry rcp server returned a nil result in violation of the rcp standard")
	}

	return *result.Result, nil

}

func debugPrintf(format string, a ...string) {
	if verbose {
		fmt.Printf(format, a)
	}
}
