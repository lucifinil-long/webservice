package webservice

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// FormData stores raw value of form data
type FormData struct {
	Name     string
	Data     []byte
	Filename string
}

// MultiFormData defines multi form data object
type MultiFormData struct {
	data map[string]*FormData
}

// GetDataBody get post/put/patch data
func GetDataBody(r *http.Request, logger Logger) (data []byte, err error) {
	logger.Debug("entered...")
	defer func() { logger.Debug("done with error", err) }()

	if !(strings.EqualFold(r.Method, "POST") ||
		strings.EqualFold(r.Method, "PUT") ||
		strings.EqualFold(r.Method, "PATCH")) {
		return nil, ErrorWrongMethod
	}

	if r.Form != nil {
		return nil, ErrorParsedYet
	}

	data, err = ioutil.ReadAll(r.Body)
	r.ParseForm()
	return
}

// GetMultiFormData reads multi form data of request
func GetMultiFormData(r *http.Request, logger Logger) (mfd *MultiFormData, err error) {
	logger.Debug("entered...")
	defer func() { logger.Debug("done with error", err) }()

	mr, err := r.MultipartReader()
	if err != nil {
		logger.Warn("r.MultipartReader returned error", err)
		return nil, err
	}

	data := make(map[string]*FormData, 16)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		// we get ipa file
		content, err := ioutil.ReadAll(p)
		if err != nil {
			logger.Warn("ioutil.ReadAll failed with", err)
			return nil, err
		}
		record := &FormData{
			Name:     p.FormName(),
			Filename: p.FileName(),
			Data:     content,
		}
		data[p.FormName()] = record
		logger.Debug("read form part", "FormName", p.FormName(), "FileName", p.FileName())
	}

	return &MultiFormData{data: data}, nil
}

// FormData get form date from form data set
func (mfd MultiFormData) FormData(key string) (data *FormData, err error) {
	if mfd.data == nil {
		return nil, ErrorNotFound
	}

	data, ok := mfd.data[key]
	if ok && data != nil {
		return data, nil
	}

	for k, v := range mfd.data {
		if strings.EqualFold(k, key) && v != nil {
			return v, nil
		}
	}

	return nil, ErrorNotFound
}

// FormDataMD5 computes multi form data md5
func (mfd MultiFormData) FormDataMD5() (MD5 string) {
	// logger.Debug("enter...")
	// defer func() { logger.Debug("computed md5", MD5) }()
	keys := make([]string, 0, len(mfd.data))
	for k := range mfd.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	encoder := md5.New()
	for _, k := range keys {
		fd := mfd.data[k]
		encoder.Write(fd.Data)
	}

	return hex.EncodeToString(encoder.Sum(nil))
}
