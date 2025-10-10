import threading
import requests
from PIL import Image
import io
import websocket
from typing import Any

# WebSocket 服务地址
WS_URL = "ws://localhost:8081/ws"


def on_message(ws: Any, message: str) -> None:
    # 按第一个空格分段
    if " " in message:
        prefix, content = message.split(" ", 1)
    else:
        prefix, content = message, ""

    print(f"[{prefix}] {content}")

    if prefix == "QrCodeId":
        # 展示二维码
        uuid = content.strip()
        qr_url = f"https://ids.xmu.edu.cn/authserver/qrCode/getCode?uuid={uuid}"
        resp = requests.get(
            qr_url,
            headers={
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0"
            },
        )
        if resp.status_code == 200:
            img = Image.open(io.BytesIO(resp.content))
            img.show()
        else:
            print(f"获取二维码失败: {resp.status_code}")
    elif prefix == "Error":
        print("错误:", content)
    elif prefix == "Session":
        print("登录成功，Session:", content)
        ws.close()


def on_error(ws: Any, error: Any) -> None:
    print("WebSocket错误:", error)


def on_close(ws: Any, close_status_code: int, close_msg: str) -> None:
    print("WebSocket连接关闭")


def on_open(ws: Any) -> None:
    # 连接建立后发送登录命令
    def run():
        ws.send("login_lnt_qr")

    threading.Thread(target=run).start()


if __name__ == "__main__":
    ws = websocket.WebSocketApp(
        WS_URL,
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )

    ws.run_forever()  # type:ignore
