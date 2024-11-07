const ConnForm = {
    conns: [],
    DBerror: "",
    connecting: false,
    oninit: () => {
        m.request({
            method: "GET",
            url: "/api/config/cnxnames",
        }).then(function (response) {
            ConnForm.conns = response.data;
        });
    },
    submitConnect: (conn) => {
        ConnForm.connecting = true;
        m.request({
            method: "GET",
            url: "/api/connect/:conn",
            params: { conn: conn },
            headers: getRequestHeaders(),
            extract: getRequestExtract(),
        }).then(function (response) {
            ConnForm.connecting = false;
            if (response.DBerror !== "") {
                ConnForm.DBerror = response.DBerror;
            }
            else {
                App.conn = conn;
                ConnInfos.get();
                SchemaForm.get();
                DataDict.getTables();
                DataDict.getViews();
                DataDict.getProcedures();
            }
        });
    },
    saveToLocalStorage: (k, v) => {
        localStorage.setItem(App.conn + ":" + k, v);
    },
    getFromLocalStorage: (k) => {
        return localStorage.getItem(App.conn + ":" + k);
    },
    view: () => {
        return !ConnForm.conns ? null : [
            m("select", {
                id: "connSelect",
                onchange: function (e) {
                    App.conn = "";
                    App.schema = "";
                    ConnForm.DBerror = "";

                    SchemaForm.reset();
                    ConnInfos.reset();
                    QryForm.reset();
                    DataDict.reset();

                    if (e.target.value === "") {
                        return;
                    }

                    ConnForm.submitConnect(e.target.value);
                }
            }, [
                m("option", { value: "" }, "select connection…"),
                m(OptGroup, { rows: ConnForm.conns })
            ]),
            ConnForm.DBerror == "" ? null : [
                m("button.ml-10", { onclick: () => { document.querySelector("#connSelect").dispatchEvent(new Event("change")); } }, "retry to connect")
            ],
            ConnForm.connecting ? m(WaitingAnimation, { text: "connecting", class: "ml-10" }) : null 
        ];
    }
}