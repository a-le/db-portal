function FileFormatInput() {
    return {
        view: function(vnode) {
            const { onChange, namePrefix = "", value = "" } = vnode.attrs || {};
            const name = namePrefix ? `${namePrefix}[format]` : "file-format";
            const options = [
                { value: "", label: "Select format…" },
                { value: "xlsx", label: "xlsx" },
                { value: "csv", label: "csv" },
                { value: "json", label: "json" },
                { value: "jsonTabular", label: "json tabular" }
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
