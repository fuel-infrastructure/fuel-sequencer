import base64
import hashlib

# From e.g. https://rpc-seq.simplystaking.xyz/block?height=1000
base64_tx_data = "CmAKXgokL2Z1ZWxzZXF1ZW5jZXIuYnJpZGdlLk1zZ1N1cHBseURlbHRhEjYKNGZ1ZWxzZXF1ZW5jZXIxMGQwN3kyNjVnbW11dnQ0ejB3OWF3ODgwam5zcjcwMGpkamZ2azMSAhIA"

bz = base64.standard_b64decode(base64_tx_data)
tx_hash = hashlib.sha256(bz).hexdigest().upper()

print(tx_hash)
