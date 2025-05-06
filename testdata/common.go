package testdata

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func ThisDir() string {
	_, b, _, _ := runtime.Caller(0)
	pwd, _ := filepath.Abs(filepath.Dir(b))

	return pwd
}

func SearchRequests() (<-chan string, error) {
	fh, err := os.Open(filepath.Join(ThisDir(), "search_requests.raw"))
	if err != nil {
		return nil, err
	}
	scan := bufio.NewScanner(fh)
	scan.Split(bufio.ScanLines)

	res := make(chan string)

	go func(r chan string) {
		for scan.Scan() {
			r <- scan.Text()
		}
		close(res)
		defer func() {
			_ = fh.Close()
		}()
	}(res)

	return res, nil
}

func CatalogData() ([]string, []string, []string, []string, error) {
	itemNames, err := parseFile("names.raw")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	itemDesc, err := parseFile("description.raw")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	catNames, err := parseFile("category.raw")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	tms, err := parseFile("trademark.raw")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return itemNames, itemDesc, catNames, tms, nil
}

func IndexData() (io.ReadCloser, io.ReadCloser) {
	ruIdxData := `и	361
в	201
для	144
на	127
с	115
не	85
из	65
или	56
это	47
а	46
от	43
как	36
см	33
при	31
по	30
к	28
его	27
до	27
поможет	24
цвет	24
лет	24
х	24
`
	enIdxData := `name	23
the	22
of	20
orient	15
swarovski	15
and	12
`

	return io.NopCloser(bytes.NewBufferString(ruIdxData)), io.NopCloser(bytes.NewBufferString(enIdxData))
}

func parseFile(path string) ([]string, error) {
	fh, err := os.Open(filepath.Join(ThisDir(), path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fh.Close()
	}()

	scan := bufio.NewScanner(fh)
	scan.Split(bufio.ScanLines)
	res := make([]string, 0, 100)
	for scan.Scan() {
		res = append(res, scan.Text())
	}

	return res, nil
}
