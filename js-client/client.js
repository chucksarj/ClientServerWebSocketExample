var i = 1;

window.addEventListener("load", function (evt) {
    var output = document.getElementById("output");
    var apName = document.getElementById("apName");
    var sensorId = document.getElementById("sensorId");
    var sensorValue = document.getElementById("sensorValue");

    var ws;
    var print = function (message) {
        var d = document.createElement("div");
        d.innerText = message;
        output.appendChild(d);
    };

    // create a connection to server
    document.getElementById("open").onclick = function (evt) {
        if (ws) {
            return false;
        }

        // open a new websocket
        ws = new WebSocket("ws://localhost:8081/ws");
        ws.binaryType = 'arraybuffer';
        ws.onopen = function (evt) {
            print("Connection to server Opened");
        }
        ws.onclose = function (evt) {
            print("Connection to server Closed");
            ws = null;
        }
        ws.onmessage = function (evt) {
            let arrayBuffer = event.data;
            let message = proto.main.SensorData.deserializeBinary(arrayBuffer);

            //print("RECIEVED: " + message.toString());
            
            print(strSdata(message));
        }
        ws.onerror = function (evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    // send message to server 
    document.getElementById("send").onclick = function (evt) {
        if (!ws) {
            return false;
        }

        let message = new proto.main.SensorData();
        message.setApName(apName.value);
        message.setSensorId(sensorId.value);
        message.setSensorValue(sensorValue.value);

        //print("Sent: " + message.toString());
        ws.send(message.serializeBinary(), { binary: true });
        return false;
    };

    // exit connection 
    document.getElementById("close").onclick = function (evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});

/**
 * @param {proto.Person} person - The person
 */
function strSdata(sdata){
    let result = "";
    result += "Record #" + i + " Ap_Name: " + sdata.getApName() + " Status: " + sdata.getStatus() + "\n";
    i += 1

    return result;
}
