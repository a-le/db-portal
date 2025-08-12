function EndpointTypeInput() {
    return {
        view: function(vnode) {
            const { onChange, endPointType = "", value = "" } = vnode.attrs || {};
            const name = endPointType ? `${endPointType}[type]` : "endpoint-type-select";
            const options = [
                { value: "", label: "select type…" },
                { value: "table", label: "DB table" },
                ...(endPointType === "origin" ? [{ value: "query", label: "SQL query" }] : []),
                { value: "file", label: "File" }
            ];
            return m(SelectInput(), {
                name,
                value,
                options,
                onchange: (e) => {
                    if (onChange) onChange(e.target.value);
                }
            });
        }
    };
}
