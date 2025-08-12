const QueryPage = {
    dsName: "",
    schema: "",
    tabState: new UIState({def: "result"}),
    SchemaInput: SchemaInput(),
    view: () => {
        return [
            m("section.area-main-menu",
                m("div.grid", { style: "grid-template-columns: auto 1fr;" },
                    m("div.grid-col",
                        m("div",
                            m(DataSourceInput, {
                                value: QueryPage.dsName,
                                onConnect: (dsName) => {
                                    QueryPage.dsName = dsName;
                                    DSInfoSection.get();
                                    QueryPage.SchemaInput.getSchemas(QueryPage.dsName, QueryPage.schema);
                                    DataDictForm.getTables();
                                    DataDictForm.getViews();
                                    DataDictForm.getProcedures();
                                },
                                onChange: () => {
                                    QueryPage.dsName = "";
                                    QueryPage.schema = "";
                                    QueryPage.SchemaInput.reset();
                                    DSInfoSection.reset();
                                    QryForm.reset();
                                    DataDictForm.reset();
                                }
                            })
                        ),
                        m("div.ml-10",
                            m(QueryPage.SchemaInput, {
                                dsName: QueryPage.dsName,
                                value: QueryPage.schema,
                                onChange: (newSchema) => {
                                    QueryPage.schema = newSchema;
                                    QryForm.reset();
                                    QryExplainForm.reset();
                                    DataDictForm.reset();
                                    DSInfoSection.get();
                                    DataDictForm.getTables();
                                    DataDictForm.getViews();
                                    DataDictForm.getProcedures();
                                }
                            })
                        ),
                        m("div.grid-col.align-items-end.ml-10", m(DSInfoSection)),
                    ),
                ),
                m("section.area-main-content",
                    !QueryPage.dsName ? null :
                        [
                            m("div.grid-query", {
                                oncreate: function (vnode) {
                                    var h0 = document.querySelector('.grid-query').offsetHeight,
                                        h1 = 195, // ideal area-query-editor height, can be greater
                                        h2 = 335, // ideal area-query-output for 15 lines of results
                                        h1 = Math.max(h0 - h2, h1);
                                    const LayoutGrid = GridResize('.grid-query', '.area-query-splitter', '.area-query-editor', `${h1}px 3px auto auto`, 195, false);
                                    LayoutGrid.init();
                                }
                            },
                                m("section.area-query-editor",
                                    m("div.grid-q-editor-datadict", {
                                        oncreate: function (vnode) {
                                            const LayoutGrid = GridResize('.grid-q-editor-datadict', '.area-q-splitter', '.area-q-editor', `1fr 3px 1fr`, 540, true);
                                            LayoutGrid.init();
                                        }
                                    },
                                        m("section.area-q-editor", m(QryForm)),
                                        m("div.area-q-splitter.splitter.splitter-vertical"),
                                        m("section.area-q-datadict", m(DataDictForm)),
                                    )
                                ),
                                m("div.area-query-splitter.splitter.splitter-horizontal"),
                                m("div",
                                    m("section.area-query-output-menu",
                                        m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                                            m("div.grid-col.tab.tab-b", {
                                                class: QueryPage.tabState.selectedClass("result"),
                                                onclick: () => QueryPage.tabState.set("result")
                                            }, "result"),
                                            m("div.grid-col.tab.tab-b.ml-20", {
                                                class: QueryPage.tabState.selectedClass("explain"),
                                                onclick: () => QueryPage.tabState.set("explain")
                                            }, "explain"),
                                            m("div.grid-col.align-items-end.ml-50", m(QryInfosSection)),
                                        ),
                                    ),
                                    m("section.area-query-output",
                                        m("div", { class: QueryPage.tabState.displayClass("result") }, m(QryResultSection)),
                                        m("div", { class: QueryPage.tabState.displayClass("explain") }, m(QryExplainForm)),
                                    )
                                )
                            )
                        ]
                )
            )
        ]
    }
};