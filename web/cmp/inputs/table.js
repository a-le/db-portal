function TableInput() {
    return {
        tables: null,
        reset: function () {
            this.tables = null;
        },
        getTables: function (dsName, schema) {
            if (dsName === "")
                return this.reset();

            const params = { dsName, command: "tables" };
            if (schema && schema !== "") 
                params.schema = schema;

            m.request({
                method: "GET",
                url: schema ? "/api/command/:dsName/:schema/:command" : "/api/command/:dsName/:command",
                headers: App.getAuthHeaders(),
                params
            }).then((response) => {
                this.tables = response.data;
            });
        },
        view: function (vnode) {
            if (!this.tables || !this.tables.rows.length)
                return null;

            const { onChange, namePrefix, value = "" } = vnode.attrs || {};
            const name = namePrefix ? `${namePrefix}[table]` : "table-select";
            const options = [
                { value: "", label: "select table…" },
                ...this.tables.rows.map(row => ({
                    value: row[0],
                    label: row[0]
                }))
            ];
            return m(SelectInput, {
                name,
                value,
                options,
                title: "set a table.",
                onchange: (e) => {
                    if (onChange) onChange(e.target.value);
                }
            });
        }
    };
}
