function SchemaInput() {
    return {
        schemas: null,
        reset: function () {
            this.schemas = null;
        },
        getSchemas: function (dsName, schema) {
            if ( dsName === "" ) 
                return this.reset();
            
            const params = { dsName, command: "schemas" };
            if (schema !== "") {
                params.schema = schema;
    }
            m.request({
                method: "GET",
                url: schema ? "/api/command/:dsName/:schema/:command" : "/api/command/:dsName/:command",
                headers: App.getAuthHeaders(),
                params,
            }).then((response) => {
                this.schemas = response.data;
            });
        },
        view: function (vnode) {
            if (!this.schemas || !this.schemas.rows.length)
                return null;

            const { schema, onChange, namePrefix, value = "" } = vnode.attrs || {};
            const name = namePrefix ? `${namePrefix}[schema]` : "schema-select";
            return m("select", {
                name,
                value,
                title: "optionnaly set a schema or database…",
                onchange: (e) => {
                    if (onChange) onChange(e.target.value);
                }
            }, [
                m("option", { value: "" }, ""),
                this.schemas.rows.map(row =>
                    m("option", { value: row[0] }, row[0])
                )
            ]);
        }
    };
}
