import csv


def write_server_cfg(record: dict):
    cfg = f"""[DEFAULT]
SocketAcceptPort=9008
ResetOnLogon=Y
FileLogPath=log

[SESSION]
BeginString=FIX.4.2
SenderCompID={record['TargetCompID']}
TargetCompID={record['SenderCompID']}
"""
    filename = f'clients/test/multi/{record["Router"]}_{record["AccountID"]}.cfg'
    with open(filename, "w", encoding="utf8") as fout:
        fout.write(cfg)
        print(f"{filename} created")


def write_client_cfg(record: dict):
    cfg = f"""[DEFAULT]
ConnectionType=initiator
FileLogPath=./log
HeartBtInt=30
UseDataDictionary=N

[SESSION]
BeginString=FIX.4.2
SenderCompID={record['SenderCompID']}
TargetCompID={record['TargetCompID']}
SocketConnectPort=9008
SocketConnectHost=113.28.233.99
# custom fields
AccountID={record['AccountID']}
"""
    filename = f'clients/multi/{record["Router"]}_{record["AccountID"]}.cfg'
    with open(filename, "w", encoding="utf8") as fout:
        fout.write(cfg)
        print(f"{filename} created")


if __name__ == "__main__":
    with open("input/multi-clients.csv", "r", encoding="utf8") as file:
        reader = csv.DictReader(file)
        for record in reader:
            write_client_cfg(record)
