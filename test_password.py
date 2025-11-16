import threading
import websocket
from typing import Any

# WebSocket 服务地址
WS_URL = "ws://localhost:8081/ws"

# 替换成实际账号和密码
USERNAME = input("请输入用户名: ").strip()
PASSWORD = input("请输入密码: ").strip()


def on_message(ws: Any, message: str) -> None:
    if " " in message:
        prefix, content = message.split(" ", 1)
    else:
        prefix, content = message, ""

    print(f"[{prefix}] {content}")

    if prefix == "Session":
        print("登录成功，Session:", content)
        ws.close()
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


# 测试字段列表
REQUIRED_FIELDS = [
    "username",
    "password",
    "name",
    "avatar",
    "job",
    "organization",
    "location",
    "email",
    "introduction",
    "personalWebsite",
    "jobName",
    "organizationName",
    "locationName",
    "phone",
    "registrationDate",
    "accountId",
    "certification",
    "role",
    "updateTime",
]


def test_profile_with_session(session: str) -> None:
    """使用 session 调用 profile 并验证返回字段"""

    def on_message(ws: Any, message: str) -> None:
        if " " in message:
            prefix, content = message.split(" ", 1)
        else:
            prefix, content = message, ""
        if prefix == "Profile":
            import json

            data = json.loads(content)
            missing = [f for f in REQUIRED_FIELDS if f not in data]
            if missing:
                print("缺少字段:", missing)
            else:
                print("所有字段均存在")
            ws.close()
        elif prefix == "Error":
            print("错误:", content)
            ws.close()

    def on_open(ws: Any) -> None:
        ws.send(f"profile {session}")

    ws = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws.run_forever()  # type: ignore


def run_password_login_and_test_profile() -> None:
    """密码登录并测试 profile"""

    def on_message(ws: Any, message: str) -> None:
        if " " in message:
            prefix, content = message.split(" ", 1)
        else:
            prefix, content = message, ""
        if prefix == "Session":
            print("登录成功，Session:", content)
            ws.close()
            test_profile_with_session(content)
        elif prefix == "Error":
            print("错误:", content)
            ws.close()

    def on_open(ws: Any) -> None:
        ws.send(f"login_lnt_password {USERNAME} {PASSWORD}")

    ws = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws.run_forever()  # type: ignore


if __name__ == "__main__":
    # 运行密码登录测试
    run_password_login_and_test_profile()
