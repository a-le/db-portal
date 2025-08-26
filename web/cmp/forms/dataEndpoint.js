function DataEndpointForm() {
    return {
        type: "",   // source type: "table", "query", "file"
        dsName: "",
        schema: "",
        table: "",
        tableMode: "existent", // "existent" or "new".
        query: "",
        format: "", // file format
        fileObject: null,
        FileInput: FileInput(),
        SchemaInput: SchemaInput(),
        TableInput: TableInput(),

        view: function (vnode) {
            const { endPointType = "origin" } = vnode.attrs || {}

            // Check if user just click export data button from query page.
            // set type/dsName/schema accordingly
            // then get schema and table list
            if (App.dataTransferAction) {
                App.dataTransferAction = false
                this.type = "query"
                this.dsName = QueryPage.dsName
                this.schema = QueryPage.schema
                this.query = QryForm.editor.getCode().trim()
                this.SchemaInput.getSchemas(this.dsName, this.schema)
            }

            return m("table", [
                m("tr", [
                    m("th", "Type"),
                    m("td.w-full", m(EndpointTypeInput, {
                        value: this.type,
                        endPointType,
                        onChange: (sel) => {
                            // reset dependant inputs
                            this.format = "";
                            this.fileObject = null;
                            if (["file", ""].includes(sel)) {
                                this.dsName = this.schema = "";
                                this.SchemaInput.reset()
                                this.TableInput.reset()
                            }
                            this.table = this.query = ""

                            this.type = sel;
                        }
                    }))
                ]),
                this.type === "file" ?
                    m("tr", [
                        m("th", "Format"),
                        m("td", m(FileFormatInput, {
                            value: this.format,
                            namePrefix: endPointType,
                            onChange: (sel) => {
                                if (endPointType === "origin") {
                                    // reset dependant inputs
                                    this.fileObject = null
                                    this.FileInput.reset()
                                }

                                this.format = sel;
                            }
                        }))
                    ]) : null,
                this.type === "file" && endPointType === "origin" && this.format ?
                    m("tr", [
                        m("th", "File"),
                        m("td", m(this.FileInput, {
                            filename: this.fileObject ? this.fileObject.name : "",
                            namePrefix: endPointType,
                            format: this.format,
                            onChange: (file) => {
                                this.fileObject = file;
                            }
                        }))
                    ]) : null,
                this.type === "table" || this.type === "query" ?
                    [
                        m("tr", [
                            m("th", "Data source"),
                            m("td", m(DataSourceInput, {
                                value: this.dsName,
                                namePrefix: endPointType,
                                onChange: (sel) => {
                                    this.schema = this.table = "" // reset dependant inputs
                                    this.dsName = sel;
                                    this.SchemaInput.getSchemas(this.dsName, this.schema)
                                    this.TableInput.getTables(this.dsName, this.schema)
                                }
                            }))
                        ]),
                        m("tr", [
                            m("th", "Schema"),
                            m("td", m(this.SchemaInput, {
                                value: this.schema,
                                namePrefix: endPointType,
                                dsName: this.dsName,
                                onChange: (sel) => {
                                    this.table = ""; // reset dependant inputs
                                    this.schema = sel;
                                    this.TableInput.getTables(this.dsName, this.schema)
                                }
                            }))
                        ]),
                    ] : null,

                this.type === "table" ?
                    [
                        m("tr", [
                            m("th.pointer", {
                                style: {opacity: this.tableMode == "existent" ? 1 : .4},
                                onclick: (e) => {
                                    this.tableMode = "existent"
                                }
                            }, "Table"),
                            m("td",
                                this.tableMode == "existent" && m(this.TableInput, {
                                    value: this.table,
                                    namePrefix: endPointType,
                                    dsName: this.dsName,
                                    schema: this.schema,
                                    onChange: (sel) => {
                                        this.table = sel
                                    }
                                }),
                            )
                        ]),
                        endPointType == "destination" && m("tr", [
                            m("th.pointer", {
                                style: {opacity: this.tableMode == "new" ? 1 : .4},
                                title: "A new table will be created with column names and types matching the source (copy from).",
                                onclick: (e) => {
                                    this.tableMode = "new"
                                }
                            }, "New table"),
                            m("td",
                                this.tableMode == "new" && this.dsName != "" && [
                                    m('input', {
                                        oncreate: (vnode) => {
                                            vnode.dom.focus();
                                        },
                                        type: "text",
                                        autocomplete: "off",
                                        placeholder: "A new table will be created",
                                        name: `${endPointType}[table]`,
                                        value: this.table,
                                        onchange: (e) => {
                                            this.table = e.target.value;
                                        }
                                    }),
                                    m('pre.info.', 'Only DB table, DB query, or JSON tabular file as source are supported.\nOther cases are not yet implemented.'),
                                    m('input', {
                                        type: "hidden",
                                        name: `${endPointType}[isNewTable]`,
                                        value: "1"
                                    })
                                ]
                            )
                        ]),
                    ] : null,

                this.type === "query" ?
                    [
                        m("tr", [
                            m("th", "SQL Query"),
                            m("td", { style: "padding-right: 0" },
                                m('textarea', {
                                    value: this.query,
                                    name: endPointType + "[query]",
                                    onchange: (e) => {
                                        this.query = e.target.value
                                    }
                                }
                                )
                            )
                        ])
                    ] : null,

            ]);
        }
    };
}
