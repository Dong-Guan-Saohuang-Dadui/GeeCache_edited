package geecache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// PeerPicker 的 PickPeer() 方法用于根据传入的 key 选择相应节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 的 Get() 方法用于从对应 group 查找缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		//QueryEscape函数对字符串进行转码使之可以安全的用在URL查询里
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	//使用Get获取响应
	res, err := http.Get(u)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil

}
