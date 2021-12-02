package restart

const (
	wait   int32 = iota // 等待restart
	begin               // 接收到restart信号
	start               // program run
	ready               // program ready
	failed              // program failed
	end                 // 收尾往old发送graceful信号
)
