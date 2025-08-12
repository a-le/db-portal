const QryResultSection = {
    currentPage: 0,
    pageSize: 15,   // number of rows by page

    initPagination: () => {
        const totalRows = QryForm.respData && QryForm.respData.rows ? QryForm.respData.rows.length : 0;
        const totalPages = Math.ceil(totalRows / QryResultSection.pageSize);
        const startIndex = QryResultSection.currentPage * QryResultSection.pageSize;
        const endIndex = startIndex + QryResultSection.pageSize;
        const setPage = function (page) {
            if (page >= 0 && page < totalPages) {
                QryResultSection.currentPage = page;
            }
        };
        return { totalRows, totalPages, startIndex, endIndex, setPage };
    },
    view: () => {
        if (QryForm.respData) {
            /* pagination */
            var { totalPages, startIndex, endIndex, setPage } = QryResultSection.initPagination();

            /* table col widths */
            var tableDim = new TableDim();
            tableDim.setRows(QryForm.respData.rows.slice(0, 10).concat([QryForm.respData.cols])) // first 10 rows + colnames
                .setCharWidth(6.5)                                               // "UbuntuMono" .8rem width in px. See style.css.
                .setAvailableWidth(document.body.clientWidth + -30)              // brittle -30
                .setTdPadding(10 + 2)                                            // table td left and right padding + space for elipsis text-overflow. See style.css.
                .calc();
        }

        if (QryForm.executing)
            return m(WaitingAnimation, { text: "waiting for results" });

        if (QryForm.error)
            return m("div.error", "error: " + QryForm.error);

        if (QryForm.respData && QryForm.respData.DBerror)
            return m("div.error", QryForm.respData.DBerror);

        if (QryForm.respData) {
            return [
                m("div", { style: "height: " + (totalPages ? "260px" : "auto") },
                    m("table.comptext", { style: "width: " + tableDim.getTotalWidth() + "px;" }, [
                        m("thead", [
                            m("tr", [
                                QryForm.respData.cols.map(function (v, idx) {
                                    return m("th", { title: v, style: "width: " + tableDim.getColWidth(idx) + "px;" }, v);
                                })
                            ])
                        ]),
                        m("tbody", [
                            QryForm.respData.rows.slice(startIndex, endIndex).map(function (row) {
                                return m("tr", row.map(function (v, i) {
                                    return m(Cell, { val: v, type: QryForm.respData.databaseTypes[i] });
                                }));
                            })
                        ])
                    ])
                ),
                !totalPages ? null : m("div.mt-5.tac",
                    m("button", {
                        onclick: function () { setPage(QryResultSection.currentPage - 1); },
                        disabled: QryResultSection.currentPage === 0 // Disable button if on first page
                    }, "Previous"),
                    m("span.ml-10", "Page " + (QryResultSection.currentPage + 1) + " of " + Math.max(totalPages, 1)),
                    m("button.ml-10", {
                        onclick: function () { setPage(QryResultSection.currentPage + 1); },
                        disabled: QryResultSection.currentPage === totalPages - 1 // Disable button if on last page
                    }, "Next")
                )
            ]
        }

    }
}
