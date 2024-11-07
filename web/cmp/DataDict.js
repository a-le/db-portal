class Dict {
    constructor() {
        this.selected = ""; // selected item name
        this.list = null; // API response object of list of items
        this.definition = null; // API response - DDL definition of selected item
        this.columns = null; // API response - list of table columns
    }
    /**
     * Change the selected item
     * @param {string} item - The item to select.
     * @returns {boolean} - Returns true if new selection needs an API call; false otherwise.
     */
    changeItem(item) {
        if (this.selected == item)
            return false;

        this.selected = item;

        // reset
        this.definition = null; // DDL definition of selected item
        this.columns = null; // API response object of selected item

        if (item === "")
            return false;

        return true;
    }
}


const DataDict = {
    tables: new Dict(),
    views: new Dict(),
    procedures: new Dict(),
    activity: new Dict(),
    tabStates: { objects: new TabState(""), tables: new TabState("columns"), views: new TabState("columns") },

    // Reset
    reset: () => {
        DataDict.tables = new Dict();
        DataDict.views = new Dict();
        DataDict.procedures = new Dict();
        DataDict.activity = new Dict();
        DataDict.tabStates = { objects: new TabState(""), tables: new TabState("columns"), views: new TabState("columns") };
    },

    // Get m.request command options
    getCommandOptions(command, args) {
        return {
            method: "GET",
            url: "/api/command/:conn/:schema/:command",
            headers: getRequestHeaders(),
            extract: getRequestExtract(),
            params: { conn: App.conn, schema: App.schema, command: command, args: args },
        };
    },

    // Get list of objects
    getTables: () => {
        let command = "tables", args = [];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.tables.list = response;
            });
    },
    getViews: () => {
        let command = "views", args = [];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.views.list = response;
            });
    },
    getProcedures: () => {
        let command = "procedures", args = [];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.procedures.list = response;
            });
    },

    // Get select object infos 
    getTableInfos: (item) => {
        if (!DataDict.tables.changeItem(item))
            return;

        let command = "table-columns", args = [item];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.tables.columns = response;
            });

        command = "table-definition", args = [item];
        opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.tables.definition = response;
            });

    },

    getViewInfos: (item) => {
        if (!DataDict.views.changeItem(item))
            return;

        let command = "view-columns", args = [item];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.views.columns = response;
            });

        command = "view-definition", args = [item];
        opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.views.definition = response;
            });
    },

    getProcedureInfos: (item) => {
        if (!DataDict.procedures.changeItem(item))
            return;

        let command = "procedure-definition", args = [item];
        let opts = DataDict.getCommandOptions(command, args);
        m.request(opts)
            .then((response) => {
                DataDict.procedures.definition = response;
            });
    },

    getActivityInfos: () => {
        let command = "activity";
        let opts = DataDict.getCommandOptions(command, []);
        m.request(opts)
            .then((response) => {
                DataDict.activity.columns = response;
            });
    },

    view: () => {
        return m("div[id=datadict].grid", { style: "grid-template-columns: auto 1fr; padding-top: 5px;" },
            m("div.datadict",
                // select table
                !DataDict.tables.list ? null : [
                    m("div.tab.tab-l", {
                        class: DataDict.tabStates.objects.selectedClass("table"),
                        onclick: () => { DataDict.tabStates.objects.set("table") },
                    },
                        m('label', {
                            for: "tableSel",
                        }, "tables (" + DataDict.tables.list.rows.length + "): "),
                        m("select[id=tableSel]", {
                            size: 1,
                            disabled: !DataDict.tables.list.rows.length,
                            onchange: function (e) {
                                DataDict.getTableInfos(e.target.value);
                            }
                        },
                            m("option", { value: "" }, ""),
                            DataDict.tables.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // select view
                !DataDict.views.list ? null : [
                    m("div.tab.tab-l.mt-10", {
                        class: DataDict.tabStates.objects.selectedClass("view"),
                        onclick: () => { DataDict.tabStates.objects.set("view") },
                    },
                        m('label', {
                            for: "viewSel",
                        }, "views (" + DataDict.views.list.rows.length + "): "),
                        m("select[id=viewSel]", {
                            size: 1,
                            disabled: !DataDict.views.list.rows.length,
                            onchange: (e) => {
                                DataDict.getViewInfos(e.target.value);
                            }
                        },
                            m("option", { value: "" }, ""),
                            DataDict.views.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // select procedure
                !DataDict.procedures.list ? null : [
                    m("div.tab.tab-l.mt-10", {
                        class: DataDict.tabStates.objects.selectedClass("procedure"),
                        onclick: () => { DataDict.tabStates.objects.set("procedure") },
                    },
                        m('label', {
                            for: "procedureSel",
                        }, "procedures (" + DataDict.procedures.list.rows.length + "): "),
                        m("select", {
                            id: "procedureSel",
                            size: 1,
                            disabled: !DataDict.procedures.list.rows.length,
                            onchange: function (e) {
                                DataDict.getProcedureInfos(e.target.value);
                            }
                        },
                            m("option", { value: "" }, ""),
                            DataDict.procedures.list.rows.map((row) => {
                                return [
                                    m("option", { value: row[0] }, row[0])
                                ];
                            })
                        )
                    )
                ],

                // activity
                !App.conn ? null : [
                    m("div.tab.tab-l.mt-10", {
                        class: DataDict.tabStates.objects.selectedClass("activity"),
                        onclick: () => { DataDict.tabStates.objects.set("activity") },
                    },
                        m('label', {
                            for: "activityBtn",
                        }, "activity"),
                        m("button", {
                            id: "activityBtn",
                            onclick: () => { DataDict.getActivityInfos(); }
                        }, "activity")
                    )

                ]
            ),

            // data dict
            m("div[id=dataDictDef].comptext.ml-10", { style: "overflow-y: auto;" },

                // table
                !DataDict.tabStates.objects.is("table") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab", {
                            class: DataDict.tabStates.tables.selectedClass("columns"),
                            onclick: () => { DataDict.tabStates.tables.set("columns") },

                        }, "columns"),
                        m("div.grid-col.tab.ml-10", {
                            class: DataDict.tabStates.tables.selectedClass("definition"),
                            onclick: () => { DataDict.tabStates.tables.set("definition") },
                        }, "definition")
                    ),
                    m("div",
                        m("div", { class: DataDict.tabStates.tables.displayClass("columns") },
                            m(DictColumns, { resp: DataDict.tables.columns, selected: DataDict.tables.selected })
                        ),
                        m("div", { class: DataDict.tabStates.tables.displayClass("definition") },
                            m(DictCode, { resp: DataDict.tables.definition, selected: DataDict.tables.selected })
                        )
                    )
                ],

                // view
                !DataDict.tabStates.objects.is("view") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab", {
                            class: DataDict.tabStates.views.selectedClass("columns"),
                            onclick: () => { DataDict.tabStates.views.set("columns") },
                        }, "columns"),
                        m("div.grid-col.tab.ml-10", {
                            class: DataDict.tabStates.views.selectedClass("definition"),
                            onclick: () => { DataDict.tabStates.views.set("definition") },
                        }, "definition")
                    ),
                    m("div",
                        m("div", { class: DataDict.tabStates.views.displayClass("columns") },
                            m(DictColumns, { resp: DataDict.views.columns, selected: DataDict.views.selected })
                        ),
                        m("div", { class: DataDict.tabStates.views.displayClass("definition") },
                            m(DictCode, { resp: DataDict.views.definition, selected: DataDict.views.selected })
                        )
                    )
                ],

                // procedure
                !DataDict.tabStates.objects.is("procedure") ? null : [
                    m("div.grid", { style: "grid-template-columns: auto auto 1fr;" },
                        m("div.grid-col.tab.selected", "definition"),
                    ),
                    m("div",
                        m(DictCode, { resp: DataDict.procedures.definition, selected: DataDict.procedures.selected })
                    )
                ],

                // activity
                !DataDict.tabStates.objects.is("activity") ? null : [
                    m("div",
                        m(DictColumns, { resp: DataDict.activity.columns, selected: "activity" })
                    )
                ]

            )

        )
    }
}