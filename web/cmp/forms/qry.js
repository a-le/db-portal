const QryForm = {
    query: "",
    respData: null,
    exportType: "",
    resizeObserver: null,
    editor: null,
    editorTheme: "",
    xhr: null,
    executing: false,
    error: false,
    selectedFileName: "",
    reset: () => {
        QryForm.query = ""
        QryForm.respData = null
        QryForm.currentPage = 0

        QryExplainForm.reset()
        QryInfosSection.reset()
    },
    // execute query and explain query forms
    submitQuery: () => {
        QryForm.respData = null
        QryForm.currentPage = 0
        QryForm.error = null
        QryForm.query = QryForm.editor.getCode().trim()
        localStorage.setItem(["lastQuery", QueryPage.dsName].join("::"), QryForm.query)
        if (!QryForm.query.length) {
            return
        }

        QryForm.executing = true
        let url, params
        params = { dsname: QueryPage.dsName }
        if (QueryPage.schema !== "") {
            url = "/api/query/:dsname/:schema"
            params.schema = QueryPage.schema
        } else {
            url = "/api/query/:dsname"
        }
        const formData = new FormData()
        formData.set("query", QryForm.query)

        m.request({
            method: "POST",
            url,
            params,
            headers: App.getAuthHeaders(),
            body: formData,
        }).then((response) => {
            QryForm.executing = false
            QryResultSection.currentPage = 0
            QryForm.respData = response.data
            QryForm.respData.duration = Math.ceil(QryForm.respData.duration / 1e+6) // nanoseconds to milliseconds
        }).catch((e) => {
            QryForm.executing = false
            QryForm.error = e.response.error;
        })
    },
    // download results form: see view
    view: () => {
        return [
            QueryPage.dsName.length &&
                [
                    m("code[id=query-code]", {
                        onclick: () => {
                            QryForm.editor.setFocusInitial()
                        },
                        oninit: (vnode) => {
                            var qryFormMenuHeight = 58 // #qryFormMenu height
                            var datadictMgBtm = 2
                            QryForm.resizeObserver = new ResizeObserver(entries => {
                                vnode.dom.style.height = entries[0].contentRect.height - qryFormMenuHeight + 'px'

                                // adjust height of area-q-datadict 1st child 
                                document.querySelector('section.area-q-datadict > :first-child').style.height = entries[0].contentRect.height - datadictMgBtm + 'px'
                            })
                        },
                        oncreate: (vnode) => {
                            const lastQuery = localStorage.getItem(["lastQuery", QueryPage.dsName].join("::")) || ''

                            QryForm.editorTheme = App.theme
                            QryForm.editor = new SqlEditor(vnode.dom.id, isLightTheme(App.theme) ? 'light' : 'dark')
                            QryForm.editor.setCode(lastQuery)
                            QryForm.editor.setFocusInitial()

                            // Start observing the element
                            QryForm.resizeObserver.observe(document.querySelector('.area-query-editor'))
                        },
                        onbeforeupdate: () => {
                            if (QryForm.editorTheme !== App.theme) {
                                if (isLightTheme(App.theme)) QryForm.editor.setLightTheme()
                                else QryForm.editor.setDarkTheme()
                                QryForm.editorTheme = App.theme
                            }
                            return false
                        },
                        onremove: () => {
                            QryForm.resizeObserver.disconnect()
                        }
                    }),
                    m("div[id=qryFormMenu]", { style: "padding: 0 6px" },
                        m("fieldset",
                            m("legend", "query execution"),
                            m("button[type=button]", {
                                disabled: QryForm.executing,
                                onclick: () => {
                                    QryExplainForm.submit()
                                    QueryPage.tabState.set("explain")
                                }
                            }, "explain"),
                            m("button[type=button].ml-10", {
                                disabled: QryForm.executing,
                                onclick: () => {
                                    QryExplainForm.reset()
                                    QryForm.submitQuery()
                                    QueryPage.tabState.set("result")
                                }
                            }, "run query"),
                            m("button[type=button]", {
                                title: "Abort execution.",
                                disabled: !QryForm.executing,
                                onclick: () => {
                                    QryForm.xhr.abort()
                                    QryForm.xhr = null
                                    QryForm.executing = false
                                }
                            }, "■"),
                        ),
                        m("fieldset", { style: "float: right" },
                            m("legend", m.trust("&#8644 copy data")),
                            m("button[type=button]", {
                                title: "Navigate to the copy data panel with current settings.",
                                disabled: QryForm.executing,
                                onclick: () => {
                                    App.dataTransferAction = true
                                    App.pageState.set("copy")
                                }
                            }, "set as source"),
                        ),

                    ),
                ]
        ]
    }
}
