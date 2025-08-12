function DataEndpointForm() {
    return {
        type: "",   // source type: table, query, file
        dsName: "",
        schema: "",
        table: "",
        query: "",
        format: "", // file format
        fileObject: null,
        FileInput: FileInput(),
        SchemaInput: SchemaInput(),
        TableInput: TableInput(),

        view: function (vnode) {
            const { endPointType = "origin" } = vnode.attrs || {};

            // Check if user just click export data button from query page.
            // set type/dsName/schema accordingly
            // then get schema and table list
            if (App.dataTransferAction) {
                App.dataTransferAction = false;
                this.type = "query";
                this.dsName = QueryPage.dsName;
                this.schema = QueryPage.schema;
                this.query = QryForm.editor.getCode().trim();
                this.SchemaInput.getSchemas(this.dsName, this.schema);
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
                            if ( ["file", ""].includes(sel) ) {
                                this.dsName = this.schema = "";
                                this.SchemaInput.reset();
                                this.TableInput.reset();
                            }
                            this.table = this.query = "";

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
                                if ( endPointType === "origin" ) {
                                    // reset dependant inputs
                                    this.fileObject = null;
                                    this.FileInput.reset();
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
                                    this.schema = this.table = ""; // reset dependant inputs
                                    this.dsName = sel;
                                    this.SchemaInput.getSchemas(this.dsName, this.schema);
                                    this.TableInput.getTables(this.dsName, this.schema);
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
                                    this.TableInput.getTables(this.dsName, this.schema);
                                }
                            }))
                        ]),
                    ] : null,

                this.type === "table" ?
                    [
                        m("tr", [
                            m("th", "Table"),
                            m("td", m(this.TableInput, {
                                value: this.table,
                                namePrefix: endPointType,
                                dsName: this.dsName,
                                schema: this.schema,
                                onChange: (sel) => {
                                    this.table = sel;
                                }
                            }))
                        ])
                    ] : null,

                this.type === "query" ?
                    [
                        m("tr", [
                            m("th", "SQL Query"),
                            m("td", {style: "padding-right: 0"}, 
                                m('textarea', {
                                    value: this.query,
                                    name: endPointType + "[query]",
                                    onChange: (value) => {
                                        this.query = value;
                                    }}
                                )
                            )
                        ])
                    ] : null,

            ]);
        }
    };
}
