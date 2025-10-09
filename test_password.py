import threading
import websocket
from typing import Any

# WebSocket 服务地址
WS_URL = "ws://localhost:8081/ws"

# 替换成实际账号和密码
USERNAME = input("请输入用户名: ")
PASSWORD = input("请输入密码: ")


def on_message(ws: Any, message: str) -> None:
    if " " in message:
        prefix, content = message.split(" ", 1)
    else:
        prefix, content = message, ""

    print(f"[{prefix}] {content}")

    if prefix == "Session":
        print("登录成功，Session:", content)
    elif prefix == "Error":
        print("错误:", content)


def on_error(ws: Any, error: Any) -> None:
    print("WebSocket错误:", error)


def on_close(ws: Any, close_status_code: int, close_msg: str) -> None:
    print("WebSocket连接关闭")


def on_open(ws: Any) -> None:
    def run():
        # 按 main.go 的协议发送 login_lnt_password + 用户名 + 密码
        ws.send(f"login_lnt_password {USERNAME} {PASSWORD}")

    threading.Thread(target=run).start()


if __name__ == "__main__":
    ws = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )

    ws.run_forever()  # type: ignore
