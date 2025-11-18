import json
import threading
from typing import Any, Optional

import websocket

WS_URL = "ws://localhost:8081/ws"

USERNAME = input("请输入志愿汇账号: ").strip()
PASSWORD = input("请输入志愿汇密码: ").strip()
DISPLAY_NAME = input("请输入志愿汇姓名（用于匹配 hours 查询）: ").strip()


REQUIRED_HOUR_FIELDS = ["credit_hours", "honor_hours", "total_hours"]


def split_message(message: str) -> tuple[str, str]:
    if " " in message:
        prefix, content = message.split(" ", 1)
    else:
        prefix, content = message, ""
    return prefix, content


def login_and_get_user_id(username: str, password: str) -> Optional[str]:
    user_id_holder: dict[str, Optional[str]] = {"value": None}

    def on_message(ws: Any, message: str) -> None:
        prefix, content = split_message(message)
        print(f"[{prefix}] {content}")
        if prefix == "Token":
            user_id_holder["value"] = content.strip()
            print("登录成功，Token:", user_id_holder["value"])
            ws.close()
        elif prefix == "Error":
            print("登录失败:", content)
            ws.close()

    def on_open(ws: Any) -> None:
        def run() -> None:
            ws.send(f"login_zyh365 {username} {password}")

        threading.Thread(target=run).start()

    def on_error(ws: Any, error: Any) -> None:
        print("WebSocket错误:", error)

    def on_close(ws: Any, close_status_code: int, close_msg: str) -> None:
        print("login_zyh365 测试连接已关闭")

    ws_app = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws_app.run_forever()  # type: ignore[arg-type]

    return user_id_holder["value"]


def test_hours_with_user(user_id: str, name: str) -> None:
    def on_message(ws: Any, message: str) -> None:
        prefix, content = split_message(message)
        print(f"[{prefix}] {content}")
        if prefix == "Hours":
            data = json.loads(content)
            missing = [field for field in REQUIRED_HOUR_FIELDS if field not in data]
            if missing:
                print("缺少字段:", missing)
            else:
                print(
                    "全部字段存在，信用时数: {credit}, 荣誉时数: {honor}, 总时数: {total}".format(
                        credit=data["credit_hours"],
                        honor=data["honor_hours"],
                        total=data["total_hours"],
                    )
                )
            ws.close()
        elif prefix == "Error":
            print("查询失败:", content)
            ws.close()

    def on_open(ws: Any) -> None:
        def run() -> None:
            ws.send(f"hours_zyh365 {user_id} {name}")

        threading.Thread(target=run).start()

    def on_error(ws: Any, error: Any) -> None:
        print("WebSocket错误:", error)

    def on_close(ws: Any, close_status_code: int, close_msg: str) -> None:
        print("hours_zyh365 测试连接已关闭")

    ws_app = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws_app.run_forever()  # type: ignore[arg-type]


def test_hours_missing_args() -> None:
    """验证缺少参数时服务端返回错误"""

    def on_message(ws: Any, message: str) -> None:
        prefix, content = split_message(message)
        print(f"[{prefix}] {content}")
        if prefix in {"Error", "Hours"}:
            ws.close()

    def on_open(ws: Any) -> None:
        ws.send("hours_zyh365 only_user_id")

    ws_app = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
    )
    ws_app.run_forever()  # type: ignore[arg-type]


if __name__ == "__main__":
    user_id = login_and_get_user_id(USERNAME, PASSWORD)
    if not user_id:
        raise SystemExit("无法获取 Token，终止测试")

    print("\n--- 开始查询志愿服务时数 ---")
    test_hours_with_user(user_id, DISPLAY_NAME)

    print("\n--- 开始测试缺失参数场景（预期 Error 参数错误） ---")
    test_hours_missing_args()
