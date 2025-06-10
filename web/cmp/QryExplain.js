const QryExplain = {
    query: "",
    resp: null,
    reset: () => {
        QryExplain.query = null;
        QryExplain.resp = null;
    },
    submit: () => {
        QryExplain.resp = null;
        QryExplain.query = QryForm.editor.getCode().trim();
        if ( !QryExplain.query.length ) {
            return;
        }
        var formData = new FormData();
        formData.set("conn", App.conn);
        formData.set("schema", App.schema);
        formData.set("query", QryExplain.query);
        formData.set("statementType", "query");
        formData.set("explain", "1");
        m.request({
            method: "POST",
            url: "/api/query",
            credentials: "include",
            headers: getRequestHeaders(formData),
            extract: getRequestExtract(),
            body: formData,
        }).then(function (response) {
            QryForm.executing = false;
            QryExplain.resp = response;
        });
    },
    view: () => {
        return [ !QryExplain.resp ? null :
            // query results
            QryExplain.resp.DBerror !== "" ? QryExplain.resp.DBerror :
                m("table.comptext", [
                    m("thead", [
                        m("tr", [
                            QryExplain.resp.rows.length ? QryExplain.resp.cols.map(function (v) {
                                return m("th", v);
                            }) : null
                        ])
                    ]),
                    m("tbody", [
                        QryExplain.resp.rows.map(function (row) {
                            return m("tr", row.map(function (v, i) {
                                return m(Cell, { val: v, type: QryExplain.resp.databaseTypes[i] })
                            }));
                        })
                    ])
                ])
        ]
    }
}