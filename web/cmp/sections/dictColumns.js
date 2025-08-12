const DictColumnsSection = {
    tableDim: null,
    resizeObserver: null,
    rowsSample: null,
    view: (vnode) => {
        const resp = vnode.attrs.resp;
        const selected = vnode.attrs.selected;

        if (resp?.rows?.length) {
            DictColumnsSection.rowsSample = resp.rows.slice(0, 10).concat([resp.cols]);
            let availableWidth = document.querySelector('#dataDictDef').clientWidth -7;

            DictColumnsSection.tableDim = new TableDim()
                .setRows(DictColumnsSection.rowsSample)
                .setCharWidth(6.5)         // "UbuntuMono" .8rem width in px. See style.css.
                .setAvailableWidth(availableWidth)
                .setTdPadding(10)          // table td left + right padding. See style.css.
                .calc();
        }

        return [
            !resp ? null : [
                resp.DBerror ? m("div.text-warning", resp.DBerror) :
                    m("table", {
                        style: { width: (DictColumnsSection.tableDim.getTotalWidth()) + "px" },
                        oninit: () => {
                            DictColumnsSection.resizeObserver = new ResizeObserver(entries => {
                                // Use requestAnimationFrame to optimize redraw
                                window.requestAnimationFrame(() => {
                                    DictColumnsSection.tableDim.setAvailableWidth(entries[0].contentRect.width).calc();
                                    m.redraw();
                                });
                            });
                        },
                        oncreate: () => {
                            DictColumnsSection.resizeObserver.observe(document.querySelector('#dataDictDef'));
                        },
                        onremove: () => {
                            DictColumnsSection.resizeObserver.disconnect();
                        }

                    }, [
                        m("caption", selected),
                        m("thead", [
                            m("tr", [
                                resp.cols.map(function (v, idx) {
                                    return m("th", { style: "width: " + DictColumnsSection.tableDim.getColWidth(idx) + "px;" }, v);
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