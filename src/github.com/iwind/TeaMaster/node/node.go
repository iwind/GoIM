package node

type Node struct {
	Id          string // ID
	Name        string // 名称
	Description string // 描述

	IP string // 所在服务器的IP

	Health      int  // 健康度 0-100
	MaxFails    int  // 最大失败尝试次数
	FailTimeout int  // 超时时间
	Lag         int  // 网络延迟，单位为ms（毫秒）
	IsOnline    bool // 是否已上线
	IsAvailable bool // 是否可用
	Weight      int  // 权重
	IsBackup    bool // 是否为备用节点，当其余节点不可用的时候才使用

	MaxMessagesPerSecond int // 每秒支持的最大消息数

	Options map[string]interface{} // 选项
	Tags    []string               // 标签，用来将节点进行分组
	States  []State                // 节点状态，包括CPU、负载、内存等信息
}

type State struct {
}
