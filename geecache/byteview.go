package geecache

// ByteView 是外部不可修改的byte数组，为了实现Value定义Len(),ByteView 用来表示缓存值
type ByteView struct {
	b []byte
}

// Len 返回byte长度
func (bV ByteView) Len() int {
	return len(bV.b)
}

// ByteSlice 返回byte的复制
func (bV ByteView) ByteSlice() []byte {
	return cloneBytes(bV.b)
}

// String 返回byte数组字符串化结果
func (bV ByteView) String() string {
	return string(bV.b)
}

//为什么分开写？
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
