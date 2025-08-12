const DictCodeSection = {
    view: (vnode) => {
        const resp = vnode.attrs.resp;
        const selected = vnode.attrs.selected;
        let code = "";

        if (resp?.rows?.length) {
            if (resp.cols.length > 1 && resp.rows.length == 1) { // postgresql
                for (var i = 0; i < resp.cols.length; i++) {
                    if (resp.cols[i].toLowerCase().startsWith("create")) {
                        code = resp.rows[0][i];
                        break;
                    }
                }
            }
            else if ( resp.cols.length === 1 && resp.rows.length > 1 ) { // mssql
                for (var i = 0; i < resp.rows.length; i++) {
                    code += resp.rows[i][0];
                }
            }
            if (code === "") code = resp.rows[0][0];
        }

        return [
            resp?.DBerror ? m("div.text-warning.mt-10", resp.DBerror) :
                code === "" ? null :
                    m("table", [
                        m("caption", selected),
                        m("tbody", [
                            m("tr",
                                m("td",
                                    m("code", {
                                        id: "viewDef",
                                        oncreate: function (vnode) {
                                            resp.editorTheme = App.theme;
                                            resp.editor = new SqlEditor(vnode.dom.id, isLightTheme(App.theme) ? 'light' : 'dark');
                                            resp.editor.setReadOnly(true);
                                            resp.editor.setCode(code);
                                        },
                                        onbeforeupdate: function () {
                                            if (resp.editorTheme !== App.theme) {
                                                if (isLightTheme(App.theme)) resp.editor.setLightTheme();
                                                else resp.editor.setDarkTheme();
                                                resp.editorTheme = App.theme;
                                            }
                                            //resp.editor.setCode(code);
                                            return false; // prevents a diff from happening 
                                        },
                                    }, null)
                                )
                            )
                        ])
                    ])
        ]

    }
}