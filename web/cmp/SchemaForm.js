const SchemaForm = {
    schemas: null,
    reset: () => {
        SchemaForm.schemas = null;
    },
    get: () => {
        m.request({
            method: "GET",
            url: "/api/command/:conn/:schema/:command",
            params: { conn: App.conn, schema: App.schema, command: "schemas" }
        }).then(function (response) {
            SchemaForm.schemas = response;
        });
    },
    view: () => {
        return [
            !SchemaForm.schemas ? null :
                !SchemaForm.schemas.rows.length ? null :
                    m("select", {
                        name: "schema-select",
                        title: "eventually set a schema or database.",
                        onchange: function (e) {
                            App.schema = e.target.value;

                            QryForm.reset();
                            QryExplain.reset();
                            DataDict.reset();

                            ConnInfos.get();
                            DataDict.getTables();
                            DataDict.getViews();
                            DataDict.getProcedures();
                        }
                    }, [
                        m("option", { value: "" }, ""),
                        SchemaForm.schemas.rows.map(function (row) {
                            return [
                                m("option", { value: row[0], selected: App.schema === row[0] }, row[0])
                            ];
                        })
                    ])
        ];
    }
}