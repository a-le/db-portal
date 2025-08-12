const QryExplainForm = {
    query: "",
    respData: null,
    error: false,
    reset: () => {
        QryExplainForm.query = null;
        QryExplainForm.respData = null;
    },
    submit: () => {
        QryExplainForm.executing = true;
        QryExplainForm.error = null
        QryExplainForm.respData = null;
        QryExplainForm.query = QryForm.editor.getCode().trim();
        if (!QryExplainForm.query.length) {
            return;
        }

        let url, params;
        params = { dsname: QueryPage.dsName };
        if (QueryPage.schema !== "") {
            url = "/api/query/:dsname/:schema";
            params.schema = QueryPage.schema;
        } else {
            url = "/api/query/:dsname";
        }

        const formData = new FormData();
        formData.set("dsName", QueryPage.dsName);
        formData.set("schema", QueryPage.schema);
        formData.set("query", QryExplainForm.query);
        formData.set("statementType", "query");
        formData.set("explain", "1");

        m.request({
            method: "POST",
            url,
            params,
            headers: App.getAuthHeaders(),
            body: formData,
        }).then(function (response) {
            QryExplainForm.executing = false;
            QryExplainForm.respData = response.data;
        }).catch((e) => {
            QryExplainForm.executing = false
            QryExplainForm.error = e.response.error;
        });
    },
    view: () => {
        if (QryExplainForm.respData) {
            /* table col widths */
            var tableDim = new TableDim();
            tableDim.setRows(QryExplainForm.respData.rows.slice(0, 10).concat([QryExplainForm.respData.cols])) // first 10 rows + colnames
                .setCharWidth(6.5)                                               // "UbuntuMono" .8rem width in px. See style.css.
                .setAvailableWidth(document.body.clientWidth + -30)              // brittle -30
                .setTdPadding(10 + 2)                                            // table td left and right padding + space for elipsis text-overflow. See style.css.
                .calc();
        }

        if (QryExplainForm.executing)
            return m(WaitingAnimation, { text: "waiting for results" });

        if (QryExplainForm.error)
            return m("div.error", "error: " + QryExplainForm.error);

        if (QryExplainForm.respData && QryExplainForm.respData.DBerror)
            return m("div.error", QryExplainForm.respData.DBerror);

        if (QryExplainForm.respData) {
            return [
                m("table.comptext", { style: "width: " + tableDim.getTotalWidth() + "px;" }, [
                    m("thead", [
                        m("tr", [
                            QryExplainForm.respData.cols.map(function (v, idx) {
                                return m("th", { title: v, style: "width: " + tableDim.getColWidth(idx) + "px;" }, v);
                            })
                        ])
                    ]),
                    m("tbody", [
                        QryExplainForm.respData.rows.map(function (row) {
                            return m("tr", row.map(function (v, i) {
                                return m(Cell, { val: v, type: QryExplainForm.respData.databaseTypes[i] })
                            }));
                        })
                    ])
                ])
            ]
        }
    }
}