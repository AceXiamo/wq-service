package live

import "wq-service/cos"

var count int

// LoadCount
// @Description: LoadCount
func LoadCount() {
	count = cos.GetCount("wq/live/")
}

// ListRecord
// @Description: ListRecord
// @param next 		下一页标识
// @return cos.ListFileRes
func ListRecord(next string) cos.ListFileRes {
	res := cos.ListFile("wq/live/", next, 10)
	res.Size = int64(count)
	return res
}
