const DataDictForm = {
    tables: new DictInput(),
    views: new DictInput(),
    procedures: new DictInput(),
    activity: new DictInput(),
    tabStates: { objects: new UIState(), tables: new UIState({ def: "columns" }), views: new UIState({ def: "columns" }) },

    // Reset
    reset: () => {
        DataDictForm.tables = new DictInput();
        DataDictForm.views = new DictInput();
        DataDictForm.procedures = new DictInput();
        DataDictForm.activity = new DictInput();
        DataDictForm.tabStates = { objects: new UIState(), tables: new UIState({ def: "columns" }), views: new UIState({ def: "columns" }) };
    },

    // Get m.request command options
    getCommandOptions(command, args) {

        let url, params;
        params = { dsName: QueryPage.dsName, command: command };
        if (QueryPage.schema !== "")
            params.schema = QueryPage.schema;

        url = QueryPage.schema ? "/api/command/:dsName/:schema/:command" : "/api/command/:dsName/:command"
        if (args && args.length) {
            const qs = new URLSearchParams();
            args.forEach(a => qs.append("args", a));
            url += "?" + qs.toString();
        }

        return {
            method: "GET",
            url,
            headers: App.getAuthHeaders(),
            params,
        };
    },

    // Get list of objects
    getTables: () => {
        let command = "tables", args = [];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.tables.list = response.data;
            });
    },
    getViews: () => {
        let command = "views", args = [];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.views.list = response.data;
            });
    },
    getProcedures: () => {
        let command = "procedures", args = [];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.procedures.list = response.data;
            });
    },

    // Get select object infos 
    getTableInfos: (item) => {
        if (!DataDictForm.tables.changeItem(item))
            return;

        let command = "table-columns", args = [item];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.tables.columns = response.data;
            });

        command = "table-definition", args = [item];
        opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.tables.definition = response.data;
            });

    },

    getViewInfos: (item) => {
        if (!DataDictForm.views.changeItem(item))
            return;

        let command = "view-columns", args = [item];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.views.columns = response.data;
            });

        command = "view-definition", args = [item];
        opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.views.definition = response.data;
            });
    },

    getProcedureInfos: (item) => {
        if (!DataDictForm.procedures.changeItem(item))
            return;

        let command = "procedure-definition", args = [item];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.procedures.definition = response.data;
            });
    },

    getActivityInfos: () => {
        let command = "activity", args = [];
        let opts = DataDictForm.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDictForm.activity.columns = response.data;
            });
    },

    view: () => {
        return m("div[id=datadict].grid", { style: "grid-template-columns: auto 1fr;" },
            m("div.datadict",
                // select table
                !DataDictForm.tables.list ? null : [
                    m("div.tab.tab-l", {
                        class: DataDictForm.tabStates.objects.selectedClass("table"),
                        onclick: () => {
                            DataDictForm.tabStates.objects.set("table");
                        },
                    },
                        m('label', {
                            for: "tableSel",
                            title: "click to reload list",
                            onclick: () => {
                                DataDictForm.getTables();
                            }
                        }, "tables (" + DataDictForm.tables.list.rows.length + "): "),
                        m("select[id=tableSel]", {
                            size: 1,
                            disabled: !DataDictForm.tables.list.rows.length,
                            onchange: function (e) {
                                DataDictForm.getTableInfos(e.target.value);
                            }
                        },
                            m("option", { value: "" }, ""),
                            DataDictForm.tables.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // select view
                !DataDictForm.views.list ? null : [
                    m("div.tab.tab-l.mt-5", {
                        class: DataDictForm.tabStates.objects.selectedClass("view"),
                        onclick: () => { DataDictForm.tabStates.objects.set("view") },
                    },
                        m('label', {
                            for: "viewSel",
                            title: "click to reload list",
                            onclick: () => {
                                DataDictForm.getViews()
                            }
                        }, "views (" + DataDictForm.views.list.rows.length + "): "),
                        m("select[id=viewSel]", {
                            size: 1,
                            disabled: !DataDictForm.views.list.rows.length,
                            onchange: (e) => {
                                DataDictForm.getViewInfos(e.target.value);
                            },
                        },
                            m("option", { value: "" }, ""),
                            DataDictForm.views.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // select procedure
                !DataDictForm.procedures.list ? null : [
                    m("div.tab.tab-l.mt-5", {
                        class: DataDictForm.tabStates.objects.selectedClass("procedure"),
                        onclick: () => { DataDictForm.tabStates.objects.set("procedure") },
                    },
                        m('label', {
                            for: "procedureSel",
                            style: "white-space: nowrap;",
                            title: "click to reload list",
                            onclick: () => {
                                DataDictForm.getProcedures()
                            }
                        }, "procedures (" + DataDictForm.procedures.list.rows.length + "): "),
                        m("select", {
                            id: "procedureSel",
                            size: 1,
                            disabled: !DataDictForm.procedures.list.rows.length,
                            onchange: function (e) {
                                DataDictForm.getProcedureInfos(e.target.value);
                            },
                        },
                            m("option", { value: "" }, ""),
                            DataDictForm.procedures.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // activity
                !QueryPage.dsName ? null : [
                    m("div.tab.tab-l.mt-15", {
                        class: DataDictForm.tabStates.objects.selectedClass("activity"),
                        onclick: () => { DataDictForm.tabStates.objects.set("activity") },
                    },
                        m("button", {
                            onclick: () => { DataDictForm.getActivityInfos(); }
                        }, "activity")
                    )

                ]
            ),

            // data dict
            m("div[id=dataDictDef].comptext.ml-10", { style: "overflow-y: auto;" },

                // table
                !DataDictForm.tabStates.objects.is("table") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab", {
                            class: DataDictForm.tabStates.tables.selectedClass("columns"),
                            onclick: () => { DataDictForm.tabStates.tables.set("columns") },

                        }, "columns"),
                        m("div.grid-col.tab.ml-10", {
                            class: DataDictForm.tabStates.tables.selectedClass("definition"),
                            onclick: () => { DataDictForm.tabStates.tables.set("definition") },
                        }, "definition")
                    ),
                    m("div",
                        m("div", { class: DataDictForm.tabStates.tables.displayClass("columns") },
                            m(DictColumnsSection, { resp: DataDictForm.tables.columns, selected: DataDictForm.tables.selected })
                        ),
                        m("div", { class: DataDictForm.tabStates.tables.displayClass("definition") },
                            m(DictCodeSection, { resp: DataDictForm.tables.definition, selected: DataDictForm.tables.selected })
                        )
                    )
                ],

                // view
                !DataDictForm.tabStates.objects.is("view") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab", {
                            class: DataDictForm.tabStates.views.selectedClass("columns"),
                            onclick: () => { DataDictForm.tabStates.views.set("columns") },
                        }, "columns"),
                        m("div.grid-col.tab.ml-10", {
                            class: DataDictForm.tabStates.views.selectedClass("definition"),
                            onclick: () => { DataDictForm.tabStates.views.set("definition") },
                        }, "definition")
                    ),
                    m("div",
                        m("div", { class: DataDictForm.tabStates.views.displayClass("columns") },
                            m(DictColumnsSection, { resp: DataDictForm.views.columns, selected: DataDictForm.views.selected })
                        ),
                        m("div", { class: DataDictForm.tabStates.views.displayClass("definition") },
                            m(DictCodeSection, { resp: DataDictForm.views.definition, selected: DataDictForm.views.selected })
                        )
                    )
                ],

                // procedure
                !DataDictForm.tabStates.objects.is("procedure") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab.selected", "definition"),
                    ),
                    m("div",
                        m(DictCodeSection, { resp: DataDictForm.procedures.definition, selected: DataDictForm.procedures.selected })
                    )
                ],

                // activity
                !DataDictForm.tabStates.objects.is("activity") ? null : [
                    m("div",
                        m(DictColumnsSection, { resp: DataDictForm.activity.columns, selected: "activity" })
                    )
                ]

            )

        )
    }
}