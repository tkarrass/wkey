// Package wkey provides an easy way to send keyboard input to a WifiKeyboard backend.
// See https://github.com/IvanVolosyuk/wifikeyboard
package wkey

import (
	"net/http"
	"fmt"
	"strconv"
	"io/ioutil"
	"regexp"
	"errors"
)

// TODO: complete list of key codes.
const (
	RIGHT  = 39
	LEFT   = 37
	RETURN = 13
	END    = 35
	BEGIN  = 36
	SHIFT  = 16
	DEL    = 46
)

type WKey struct {
	host string
	seq  int
}

var rxSequence *regexp.Regexp

func init() {
	// Needed for extracting the sequence number from the page source.
	rxSequence = regexp.MustCompile("seqConfirmed = ([0-9]+);")
}

// qry simplifies the http queries.
func qry(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	rdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(rdata[:]), nil
}

// Connect to a WifiKeyboard backend using the given host.
// Parameter host is the androids hostname or ip address only, without http-prefix or
// port number.
func Connect(host string) (*WKey, error) {
	resp, err := qry(fmt.Sprintf("http://%v:7777/", host))
	if err != nil {
		return nil, err
	}

	// Extract the sequence number from the page source:
	sequence := rxSequence.FindStringSubmatch(resp)[1]
	seqint, _ := strconv.Atoi(sequence)
	if seqint < 1 {
		return nil, errors.New("cannot read sequence index")
	}

	ret := &WKey{
		host: host,
		seq:  seqint,
	}
	return ret, nil
}

// Send one or more characters to the device.
// Characters will be sent as char codes - NOT as key codes.
// If you want to send key codes use the Key, KeyDown, KeyUp functions instead.
func (k *WKey) Send(s string) error {
	// assemble url
	buf := ""
	for _, r := range s {
		// yes, a buffer would be way faster, but … reasons
		buf = fmt.Sprintf("C%v,%v", r, buf)
		k.seq++
	}
	url := fmt.Sprintf("http://%v:7777/key?%v,%v", k.host, k.seq-1, buf)
	resp, err := qry(url)
	if err != nil {
		return err
	}
	if resp != "ok" {
		return errors.New(resp)
	}
	return nil
}

// Send a key down event for a given key code.
func (k *WKey) KeyDown(code int) error {
	buf := fmt.Sprintf("D%v,", code)
	url := fmt.Sprintf("http://%v:7777/key?%v,%v", k.host, k.seq, buf)
	resp, err := qry(url)
	if err != nil {
		return err
	}
	if resp != "ok" {
		return errors.New(resp)
	}
	k.seq++
	return nil
}

// Send a key up event for a given key code.
func (k *WKey) KeyUp(code int) error {
	buf := fmt.Sprintf("U%v,", code)
	url := fmt.Sprintf("http://%v:7777/key?%v,%v", k.host, k.seq, buf)
	resp, err := qry(url)
	if err != nil {
		return err
	}
	if resp != "ok" {
		return errors.New(resp)
	}
	k.seq++
	return nil
}

// Simulate a key press by sending a key down and a key up event for a given key code.
func (k *WKey) Key(code int) {
	k.KeyDown(code)
	k.KeyUp(code)
}

// Remove the current line from the text box.
func (k *WKey) Clear() {
	k.Key(END)
	k.KeyDown(SHIFT)
	k.Key(BEGIN)
	k.KeyUp(SHIFT)
	k.Key(DEL)
}
