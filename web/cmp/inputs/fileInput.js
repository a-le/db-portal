function FileInput() {
    return {
        fileInputEl: null,
        triggerFileInput: function () {
            if (this.fileInputEl) this.fileInputEl.click();
        },
        reset: function () {
            if (this.fileInputEl) {
                this.fileInputEl.value = "";
            }
        },
        view: function(vnode) {
            const { onChange, namePrefix , format, filename = "" } = vnode.attrs || {};
            const name = namePrefix ? `${namePrefix}[file]` : "endpoint-type-select";
            // Infer accept attribute from format 
            // set accept to format first 4 chars
            let accept = "";
            if (format) {
                let ext = format;
                if (ext.length > 4) ext = ext.slice(0, 4);
                accept = "." + ext;
            }
            return [
                m("button[type=button]", {
                    onclick: () => this.triggerFileInput()
                }, "Choose file..."),
                m("input[type=file]", {
                    style: "display:none",
                    name,
                    accept,
                    onchange: (e) => {
                        if (onChange) onChange(e.target.files[0], e);
                    },
                    oncreate: vnode => { this.fileInputEl = vnode.dom; },
                    title: "select a file"
                }),
                m("span.ml-10", filename),
            ];
        }
    };
}
