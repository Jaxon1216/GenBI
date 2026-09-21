package dto

// ChartAddRequest 新增图表请求（对齐前端 typings）。
type ChartAddRequest struct {
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	ChartData string `json:"chartData"`
	ChartType string `json:"chartType"`
}

// ChartEditRequest 编辑图表请求（本人）。
type ChartEditRequest struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	ChartData string `json:"chartData"`
	ChartType string `json:"chartType"`
}

// ChartUpdateRequest 更新图表请求（管理员）。
type ChartUpdateRequest struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	ChartData string `json:"chartData"`
	ChartType string `json:"chartType"`
	GenChart  string `json:"genChart"`
	GenResult string `json:"genResult"`
}

// ChartQueryRequest 图表分页查询请求。
type ChartQueryRequest struct {
	Current   int64  `json:"current"`
	PageSize  int64  `json:"pageSize"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Goal      string `json:"goal"`
	ChartType string `json:"chartType"`
	UserID    int64  `json:"userId"`
	SortField string `json:"sortField"`
	SortOrder string `json:"sortOrder"`
}

// GenChartByAiRequest AI 生成图表的表单参数（multipart 里除 file 外的字段）。
type GenChartByAiRequest struct {
	Name      string `form:"name"`
	Goal      string `form:"goal"`
	ChartType string `form:"chartType"`
}

// BiResponse AI 同步生成的返回体。
// 注意：字段名用 genChart（纠正 Java 的 genChartStr），对齐前端 AddChart 页面。
type BiResponse struct {
	ChartID   int64  `json:"chartId"`
	GenChart  string `json:"genChart"`
	GenResult string `json:"genResult"`
}
