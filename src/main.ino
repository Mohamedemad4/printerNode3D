///
/////////////////
/////////////////
/////////////////
/////////////////
//Please Test Me/
/////////////////
/////////////////
/////////////////
///
const char* ssid = "KMT House_5";
const char* password = "welovehacking5";


#include <ESP8266WiFi.h>
//how many clients should be able to connect to this ESP8266
#define MAX_SRV_CLIENTS 1
WiFiServer server(9999);
WiFiClient serverClients[MAX_SRV_CLIENTS];
void setup() {
  Serial.begin(9600);
  WiFi.begin(ssid, password);
  Serial.print("\nConnecting to "); Serial.println(ssid);
  uint8_t i = 0;
  while (WiFi.status() != WL_CONNECTED && i++ < 20) delay(500);
  if(i == 21){
    Serial.print("Could not connect to"); Serial.println(ssid);
    while(1) delay(500);
  }
  //start UART and the server
  server.begin();
  server.setNoDelay(true);
  
  Serial.print("Ready! localIP: ");
  Serial.print(WiFi.localIP());
  Serial.println("");
}
void loop() {
  uint8_t i;
  //check if there are any new clients
  if (server.hasClient()){
    for(i = 0; i < MAX_SRV_CLIENTS; i++){
      //find free/disconnected spot
      if (!serverClients[i] || !serverClients[i].connected()){
        if(serverClients[i]) serverClients[i].stop();
        serverClients[i] = server.available();
        Serial.print("Connected To Client: "); Serial.print(i);
        continue;
      }
    }
    //no free/disconnected spot so reject
    WiFiClient serverClient = server.available();
    serverClient.stop();
  }
  //check clients for data
  for(i = 0; i < MAX_SRV_CLIENTS; i++){
    if (serverClients[i] && serverClients[i].connected()){
      if(serverClients[i].available()){
        while(serverClients[i].available()){ 
            Serial.write(serverClients[i].read());
            }
        }
        if(Serial.available()){
            size_t len = Serial.available();
            uint8_t sbuf[len];
            Serial.readBytes(sbuf, len);
            //push UART data to all connected telnet clients
            for(i = 0; i < MAX_SRV_CLIENTS; i++){
              if (serverClients[i] && serverClients[i].connected()){
                serverClients[i].write(sbuf, len);
                //delay(1);
          }
        }
      }
    }
  }
}