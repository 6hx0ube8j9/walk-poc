package main

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var ni *walk.NotifyIcon
	var isExiting bool // 标记：区分是用户点击右上角 [X] 还是托盘彻底退出

	// 1. 创建基础主窗口 (原版 lxn/walk 无需 InitApp，由 Create 创建句柄)
	err := MainWindow{
		AssignTo: &mw,
		Title:    "Walk (lxn 原版) 窗口生命周期测试 (POC)",
		MinSize:  Size{Width: 400, Height: 240},
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 12},
		Children: []Widget{
			Label{
				Text: "【测试验证步骤】\n" +
					"1. 点击右上角 [X]，观察控制台是否捕获拦截日志且窗口隐藏。\n" +
					"2. 左键点击任务栏右下角托盘图标，测试窗口能否重新显示并获得焦点。\n" +
					"3. 右键托盘图标选择「彻底退出」，验证进程能否正常退出。",
			},
			PushButton{
				Text: "测试代码调用 SetVisible(false)",
				OnClicked: func() {
					mw.SetVisible(false)
					log.Println("[UI] 手动 SetVisible(false) 隐藏窗口")
				},
			},
		},
	}.Create()
	if err != nil {
		log.Fatalf("创建窗口失败: %v", err)
	}

	// 2. 【核心验证点】：原版 lxn/walk 原生 Closing 拦截
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if !isExiting {
			*canceled = true     // 撤销关闭请求
			mw.SetVisible(false) // 仅隐藏窗口
			log.Println("[Event] 成功捕获并拦截 [X] 点击，窗口已转为隐藏，保留上下文")
		} else {
			log.Println("[Event] 检测到退出标记，放行窗口销毁流程")
		}
	})

	// 3. 创建系统托盘 (原版 lxn/walk 必须显式传入宿主窗口 mw)
	ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatalf("创建托盘失败: %v", err)
	}
	defer ni.Dispose()

	ni.SetIcon(walk.IconApplication())
	ni.SetToolTip("Walk POC (lxn 原版)")
	ni.SetVisible(true)

	// 左键点击托盘：恢复显示并拉到最前
	ni.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			if !mw.Visible() {
				mw.Show()
			}
			mw.SetFocus()
			log.Println("[Tray] 左键点击托盘，窗口已恢复可见")
		}
	})

	// 托盘右键菜单
	showAction := walk.NewAction()
	showAction.SetText("显示窗口")
	showAction.Triggered().Attach(func() {
		mw.Show()
		mw.SetFocus()
		log.Println("[Tray] 菜单触发：显示窗口")
	})
	ni.ContextMenu().Actions().Add(showAction)

	exitAction := walk.NewAction()
	exitAction.SetText("彻底退出")
	exitAction.Triggered().Attach(func() {
		log.Println("[Tray] 菜单触发：正在退出...")
		isExiting = true
		mw.Close() // 触发 Closing 事件，因 isExiting 为 true，正常结束
	})
	ni.ContextMenu().Actions().Add(exitAction)

	// 4. 显示窗口并启动消息循环
	mw.Show()
	log.Println("[App] lxn 原版 POC 已启动，消息循环开始")
	mw.Run() // 原版通过 mw.Run() 进入 Windows 消息循环
	log.Println("[App] 消息循环已结束，进程退出")
}
