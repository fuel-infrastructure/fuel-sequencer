import json

import websocket  # Dependency: websocket-client

websocket_url = "ws://localhost:26657/websocket"

# Query a specific full Ethereum block sync
# query = """
# fuelsequencer.bridge.EventEthereumBlockSynced.block_number='"100"'
# AND
# fuelsequencer.bridge.EventEthereumBlockSynced.full_sync='true'
# """

# Query any full Ethereum block sync
# query = """
# fuelsequencer.bridge.EventEthereumBlockSynced.full_sync='true'
# """

# Query by message type
# query = "message.action = '/fuelsequencer.bridge.v1.MsgSupplyDelta'"

# Query for new blocks
query = "tm.event = 'NewBlock'"


def on_message(ws, message):
    response = json.loads(message)
    if 'result' in response:
        if response['result'] == {}:
            print("Received empty result - ignore this if it's the first one")
            return
        print(f"Received result: {response['result']}")
    elif 'error' in response:
        print(f"Error received: {response['error']}")


def on_error(ws, error):
    print(f"Error: {error}")


def on_close(ws, close_status_code, close_msg):
    print(f"WebSocket closed: {close_status_code}, {close_msg}")


def on_open(ws):
    print("WebSocket connection opened")
    # Subscribe to new blocks
    subscription_msg = {
        "jsonrpc": "2.0",
        "method": "subscribe",
        "id": 1,
        "params": {"query": query}
    }
    ws.send(json.dumps(subscription_msg))


def run_websocket():
    ws = websocket.WebSocketApp(websocket_url,
                                on_open=on_open,
                                on_message=on_message,
                                on_error=on_error,
                                on_close=on_close)
    ws.run_forever()


run_websocket()
