(function(G){ 'use strict';
    if (G.msg != null) {
        return;
    }

    var msg = {};
    G.msg = msg;
    var m = G.msglib;

    msg.MsgPlayer = m.MioStruct({
        playerid : m.MioField(1, m.MioInt32),
        name : m.MioField(2, m.MioString),
        uid : m.MioField(3, m.MioInt64),
        skill : m.MioField(4, m.MioString)
    });

})(this);
