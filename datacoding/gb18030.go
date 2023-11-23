package datacoding

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GB18030 编码
// gb2312、GBK、Big5 和 GB18030 之间的区别: https://github.com/bingoohuang/blog/issues/130
type GB18030 []byte

func (g GB18030) Name() DataCoding {
	return DataCodingGB18030
}

func (g GB18030) Encode() ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(g), simplifiedchinese.GB18030.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func (g GB18030) Decode() ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(g), simplifiedchinese.GB18030.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func (g GB18030) SplitBy() (maxLen, splitBy int) {
	return MaxLongSmsLength, SplitBy134
}
