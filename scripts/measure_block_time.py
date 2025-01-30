from datetime import datetime

import requests

commit_url = "http://localhost:26657/commit?height={}"

height1 = 2
height2 = "" # "" for latest

commit1 = requests.get(commit_url.format(height1)).json()
commit2 = requests.get(commit_url.format(height2)).json()

date1 = commit1['result']['signed_header']['header']['time']
date1 = date1[0:date1.index('.')+7] + 'Z'
date2 = commit2['result']['signed_header']['header']['time']
date2 = date2[0:date2.index('.')+7] + 'Z'

height1 = int(commit1['result']['signed_header']['header']['height'])
height2 = int(commit2['result']['signed_header']['header']['height'])

data1_parsed = datetime.strptime(date1, "%Y-%m-%dT%H:%M:%S.%fZ")
data2_parsed = datetime.strptime(date2, "%Y-%m-%dT%H:%M:%S.%fZ")

diff_seconds = data2_parsed.timestamp() - data1_parsed.timestamp()
seconds_per_block = diff_seconds / (height2 - height1)

print(f"Height {height1} to {height2}")
print(f"Date 1 {date1}")
print(f"Date 2 {date2}")
print(f"Time elapsed {diff_seconds} seconds")
print(f"Block time {seconds_per_block}")
