import sys
import time
import tqdm
import socket
import requests as req
import os
file=sys.argv[1]
nodeserver=sys.argv[2] #"10.10.120.81"

def SendCode(code):
    code=code+'\n'
    conn.send(code.encode())
    #while True:
    #    data=conn.recv(2000)
        #if '20' in data.decode():
    #        break


gcodelines=open(file,'r').read().split('\n')

conn=socket.socket(socket.AF_INET,socket.SOCK_STREAM)
conn.connect((nodeserver,23))
speedmode=True
commands=0
SendCode("M28 logspeed.gco")
time.sleep(2)
for i in tqdm.tqdm(gcodelines):
        SendCode(i)
        commands+=1

        if commands==80: #100,80
            commands=0
            speedmode=False
        else:
            speedmode=True

        if speedmode:
            time.sleep(0.01) #0.01
            pass
        else:
            time.sleep(2) # 5,3,20 with 80
SendCode("M29")
print("sent {0} Lines".format(gline))