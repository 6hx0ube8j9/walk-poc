package main

import (
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/tailscale/walk"
	. "github.com/tailscale/walk/declarative"
)

func main() {
	// 1. 同时输出到终端和同级目录 walk-poc.log (防止窗口秒退丢日志)
	logFile, err := os.OpenFile("walk-poc.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[FATAL PANIC] 捕获到运行时崩溃: %v\n堆栈跟踪:\n%s", r, debug.Stack())
		}
		log.Println("[App] main 函数生命周期结束，进程正式退出")
	}()

	var mw *walk.MainWindow
	var ni *walk.NotifyIcon
	var isExiting bool

	app, err := walk.InitApp()
	if err != nil {
		log.Fatalf("初始化 Walk App 失败: %v", err)
	}

	err = MainWindow{
		AssignTo: &mw,
		Title:    "Walk (tailscale 分支) 生命周期追踪",
		MinSize:  Size{Width: 400, Height: 240},
		Size:     Size{Width: 460, Height: 280},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 12},
		Children: []Widget{
			Label{
				Text: "请直接点击右上角 [X] 测试关闭拦截。\n" +
					"若窗口秒退，请直接查看同级目录生成的 walk-poc.log。",
			},
			PushButton{
				Text: "测试代码调用 SetVisible(false)",
				OnClicked: func() {
					mw.SetVisible(false)
					log.Println("[UI] 手动 SetVisible(false) 隐藏窗口成功")
				},
			},
		},
	}.Create()
	if err != nil {
		log.Fatalf("创建窗口失败: %v", err)
	}

	// 核心拦截点：逐行埋点，观察执行到了哪一步
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		log.Printf("[Event] Closing 事件已被触发! isExiting=%v, CloseReason=%d", isExiting, reason)
		if !isExiting {
			*canceled = true
			log.Println("[Event] 已将 *canceled 置为 true")
			
			mw.SetVisible(false)
			log.Println("[Event] SetVisible(false) 调用完成，窗口理论上已隐藏")
		} else {
			log.Println("[Event] 检测到退出标志，放行窗口销毁流程")
		}
	})

	ni, err = walk.NewNotifyIcon()
	if err != nil {
		log.Fatalf("创建托盘失败: %v", err)
	}
	defer ni.Dispose()

	ni.SetIcon(walk.IconApplication())
	ni.SetToolTip("Walk POC (tailscale 追踪模式)")
	ni.SetVisible(true)

	ni.MouseUp().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			log.Println("[Tray] 左键点击托盘，尝试唤醒窗口")
			if !mw.Visible() {
				mw.Show()
			}
			mw.SetFocus()
		}
	})

	exitAction := walk.NewAction()
	exitAction.SetText("彻底退出")
	exitAction.Triggered().Attach(func() {
		log.Println("[Tray] 菜单触发彻底退出")
		isExiting = true
		mw.Close()
	})
	ni.ContextMenu().Actions().Add(exitAction)

	mw.Show()
	log.Println("[App] tailscale 分支 POC 已就绪，进入 app.Run() 消息循环")
	exitCode := app.Run()
	log.Printf("[App] app.Run() 消息循环退出，返回码: %d", exitCode)
}
