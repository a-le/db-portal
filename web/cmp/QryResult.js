const QryResult = {
    currentPage: 0, // Initialize the current page to 0
    pageSize: 15,   // Set the page size to 10

    initPagination: () => {
        const totalRows = QryForm.resp && QryForm.resp.rows ? QryForm.resp.rows.length : 0;
        const totalPages = Math.ceil(totalRows / QryResult.pageSize);
        const startIndex = QryResult.currentPage * QryResult.pageSize;
        const endIndex = startIndex + QryResult.pageSize;
        const setPage = function (page) {
            if (page >= 0 && page < totalPages) {
                QryResult.currentPage = page;
            }
        };
        return { totalRows, totalPages, startIndex, endIndex, setPage };
    },
    view: () => {
        if (QryForm.resp) {
            /* pagination */
            var { totalPages, startIndex, endIndex, setPage } = QryResult.initPagination();

            /* table col widths */
            var tableDim = new TableDim();
            tableDim.setRows(QryForm.resp.rows.slice(0, 10).concat([QryForm.resp.cols])) // first 10 rows + colnames
                .setCharWidth(6.5)                                               // "UbuntuMono" .8rem width in px. See style.css.
                .setAvailableWidth(document.body.clientWidth + -30)              // brittle -30
                .setTdPadding(10 + 2)                                            // table td left and right padding + space for elipsis text-overflow. See style.css.
                .calc();
        }
        return [
            QryForm.executing ? m(WaitingAnimation, { text: "waiting for results" }) :
                QryForm.callError ? m("div.text-warning", "error: " + QryForm.callError) :
                    !QryForm.resp ? null :
                        QryForm.resp.DBerror !== "" ? QryForm.resp.DBerror : [
                            m("div.mt-5", { style: "height: " + (totalPages ? "260px" : "auto") },
                                m("table.comptext", { style: "width: " + tableDim.getTotalWidth() + "px;" }, [
                                    m("thead", [
                                        m("tr", [
                                            QryForm.resp.cols.map(function (v, idx) {
                                                return m("th", { title: v, style: "width: " + tableDim.getColWidth(idx) + "px;" }, v);
                                            })
                                        ])
                                    ]),
                                    m("tbody", [
                                        QryForm.resp.rows.slice(startIndex, endIndex).map(function (row) {
                                            return m("tr", row.map(function (v, i) {
                                                return m(Cell, { val: v, type: QryForm.resp.databaseTypes[i] });
                                            }));
                                        })
                                    ])
                                ])
                            ),
                            !totalPages ? null : m("div.mt-5.tac",
                                m("button", {
                                    onclick: function () { setPage(QryResult.currentPage - 1); },
                                    disabled: QryResult.currentPage === 0 // Disable button if on first page
                                }, "Previous"),
                                m("span.ml-10", "Page " + (QryResult.currentPage + 1) + " of " + Math.max(totalPages, 1)),
                                m("button.ml-10", {
                                    onclick: function () { setPage(QryResult.currentPage + 1); },
                                    disabled: QryResult.currentPage === totalPages - 1 // Disable button if on last page
                                }, "Next")
                            )
                        ]
        ]
    }
}
