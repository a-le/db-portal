const DictColumns = {
    tableDim: null,
    resizeObserver: null,
    rowsSample: null,
    view: (vnode) => {
        const resp = vnode.attrs.resp;
        const selected = vnode.attrs.selected;

        if (resp?.rows?.length) {
            DictColumns.rowsSample = resp.rows.slice(0, 10).concat([resp.cols]);
            let availableWidth = document.querySelector('#dataDictDef').clientWidth -7;

            DictColumns.tableDim = new TableDim()
                .setRows(DictColumns.rowsSample)
                .setCharWidth(6.5)         // "UbuntuMono" .8rem width in px. See style.css.
                .setAvailableWidth(availableWidth)
                .setTdPadding(10)          // table td left + right padding. See style.css.
                .calc();
        }

        return [
            !resp ? null : [
                resp.DBerror ? m("div.text-warning", resp.DBerror) :
                    m("table", {
                        style: { width: (DictColumns.tableDim.getTotalWidth()) + "px" },
                        oninit: () => {
                            DictColumns.resizeObserver = new ResizeObserver(entries => {
                                // Use requestAnimationFrame to optimize redraw
                                window.requestAnimationFrame(() => {
                                    DictColumns.tableDim.setAvailableWidth(entries[0].contentRect.width).calc();
                                    m.redraw();
                                });
                            });
                        },
                        oncreate: () => {
                            DictColumns.resizeObserver.observe(document.querySelector('#dataDictDef'));
                        },
                        onremove: () => {
                            DictColumns.resizeObserver.disconnect();
                        }

                    }, [
                        m("caption", selected),
                        m("thead", [
                            m("tr", [
                                resp.cols.map(function (v, idx) {
                                    return m("th", { style: "width: " + DictColumns.tableDim.getColWidth(idx) + "px;" }, v);
                                })
                            ])
                        ]),
                        m("tbody", [
                            resp.rows.map(function (row) {
                                return m("tr", row.map(function (v, i) {
                                    return m(Cell, { val: v, type: resp.databaseTypes[i] });
                                }));
                            })
                        ])
                    ])
            ]
        ]
    }
}

const DictCode = {
    view: (vnode) => {
        const resp = vnode.attrs.resp;
        const selected = vnode.attrs.selected;
        let code = "";

        if (resp?.rows?.length) {
            if (resp.cols.length > 1 && resp.rows.length == 1) { // postgresql
                for (var i = 0; i < resp.cols.length; i++) {
                    if (resp.cols[i].toLowerCase().startsWith("create")) {
                        code = resp.rows[0][i];
                        break;
                    }
                }
            }
            else if ( resp.cols.length === 1 && resp.rows.length > 1 ) { // mssql
                for (var i = 0; i < resp.rows.length; i++) {
                    code += resp.rows[i][0];
                }
            }
            if (code === "") code = resp.rows[0][0];
        }

        return [
            resp?.DBerror ? m("div.text-warning.mt-10", resp.DBerror) :
                code === "" ? null :
                    m("table", [
                        m("caption", selected),
                        m("tbody", [
                            m("tr",
                                m("td",
                                    m("code", {
                                        id: "viewDef",
                                        oncreate: function (vnode) {
                                            resp.editorTheme = App.theme;
                                            resp.editor = new SqlEditor(vnode.dom.id, isLightTheme(App.theme) ? 'light' : 'dark');
                                            resp.editor.setReadOnly(true);
                                            resp.editor.setCode(code);
                                        },
                                        onbeforeupdate: function () {
                                            if (resp.editorTheme !== App.theme) {
                                                if (isLightTheme(App.theme)) resp.editor.setLightTheme();
                                                else resp.editor.setDarkTheme();
                                                resp.editorTheme = App.theme;
                                            }
                                            //resp.editor.setCode(code);
                                            return false; // prevents a diff from happening 
                                        },
                                    }, null)
                                )
                            )
                        ])
                    ])
        ]

    }
}