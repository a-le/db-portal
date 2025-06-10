const QryForm = {
    query: "",
    resp: null,
    exportType: "",
    resizeObserver: null,
    editor: null,
    editorTheme: "",
    xhr: null,
    executing: false,
    callError: false,
    selectedFileName: "",
    reset: () => {
        QryForm.query = "";
        QryForm.resp = null;
        QryForm.currentPage = 0;

        QryExplain.reset();
        QryInfos.reset();
    },
    // execute query and explain query forms
    submitQuery: () => {
        QryForm.resp = null;
        QryForm.currentPage = 0;
        QryForm.query = QryForm.editor.getCode().trim();
        ConnForm.saveToLocalStorage('lastQuery', QryForm.query);
        if (!QryForm.query.length) {
            return;
        }
        QryForm.executing = true;
        var formData = new FormData();
        formData.set("conn", App.conn);
        formData.set("schema", App.schema);
        formData.set("query", QryForm.query);

        m.request({
            method: "POST",
            url: "/api/query",
            credentials: "include",
            headers: getRequestHeaders(formData),
            extract: getRequestExtract(),
            config: function (xhr) {
                QryForm.xhr = xhr;
            },
            body: formData,
        }).then((response) => {
            QryForm.executing = false;
            QryForm.callError = null;
            QryResult.currentPage = 0;
            QryForm.resp = response;
            QryForm.resp.duration = Math.ceil(QryForm.resp.duration / 1e+6); // nanoseconds to milliseconds
        }).catch((e) => {
            QryForm.executing = false;
            QryForm.callError = e.code ? e.code + ": " + e.message : "no error message. The server did not respond.";
        });
    },
    // download results form: see view
    view: () => {
        return ConnForm.DBerror !== "" ? m('code.text-warning', ConnForm.DBerror) :
            !App.conn.length ? null :
                [
                    m("code[id=query-code]", {
                        onclick: () => {
                            QryForm.editor.setFocusInitial();
                        },
                        oninit: (vnode) => {
                            var qryFormMenuHeight = 58; // #qryFormMenu height
                            var datadictMgBtm = 2;
                            QryForm.resizeObserver = new ResizeObserver(entries => {
                                vnode.dom.style.height = entries[0].contentRect.height - qryFormMenuHeight + 'px';

                                // adjust height of area-q-datadict 1st child 
                                document.querySelector('section.area-q-datadict > :first-child').style.height = entries[0].contentRect.height - datadictMgBtm + 'px';
                            });
                        },
                        oncreate: (vnode) => {
                            QryForm.editorTheme = App.theme;
                            QryForm.editor = new SqlEditor(vnode.dom.id, isLightTheme(App.theme) ? 'light' : 'dark');
                            QryForm.editor.setCode(ConnForm.getFromLocalStorage('lastQuery') || '');
                            QryForm.editor.setFocusInitial();
                            
                            // Start observing the element
                            QryForm.resizeObserver.observe(document.querySelector('.area-query-editor'));
                        },
                        onbeforeupdate: () => {
                            if (QryForm.editorTheme !== App.theme) {
                                if (isLightTheme(App.theme)) QryForm.editor.setLightTheme();
                                else QryForm.editor.setDarkTheme();
                                QryForm.editorTheme = App.theme;
                            }
                            return false;
                        },
                        onremove: () => {
                            QryForm.resizeObserver.disconnect();
                        }
                    }),
                    m("div[id=qryFormMenu]", { style: "padding: 0 6px;" },
                        m("fieldset",
                            m("legend", "download result"),
                            /* it uses a classic form to permit file download for best browser memory usage */
                            m("form", {
                                method: "POST",
                                action: "/api/export",
                                target: "exportpage",
                                onsubmit: (e) => {
                                    let query = QryForm.editor.getCode();
                                    e.target.elements["query"].value = query;
                                    if (query.trim() === "") return false;

                                    // Add CSRF token to form
                                    const csrfInput = document.createElement('input');
                                    csrfInput.type = 'hidden';
                                    csrfInput.name = '_csrf';
                                    csrfInput.value = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
                                    e.target.appendChild(csrfInput);

                                    return true; // let the browser continue form submission
                                }
                            },
                                m("select[name=exportType][required].w-80", { title: "choose file format to export to" },
                                    m("option", { value: "csv" }, ".csv"),
                                    m("option", { value: "json" }, ".json"),
                                    m("option", { value: "jsoncompact" }, ".json compact"),
                                    m("option", { value: "xlsx" }, ".xlsx"),
                                ),
                                m("input[type=checkbox][name=gz][id=gz].ml-10"),
                                m("label]", { for: "gz", title: "compress result with gzip" }, ".gz"),
                                m('input[name=conn][type="hidden"]', { value: App.conn }),
                                m('input[name=schema][type="hidden"]', { value: App.schema }),
                                m('input[name=query][type="hidden"]'),
                                m("button[type=submit].ml-10", {
                                    title: "execute query and download results",
                                    disabled: QryForm.executing,
                                }, m(DownloadIcon)),
                            ),
                        ),
                        // m("fieldset",
                        //     m("legend", "upload and import file"),
                        //     m("form", {
                        //         method: "POST",
                        //         action: "/api/import",
                        //         enctype: "multipart/form-data",
                        //         target: "importPopup",
                        //         onsubmit: (e) => {
                        //             // open a popup window for the form's target
                        //             let popup = window.open('', 'importPopup', 'width=600,height=400');
                        //             let content = '<html><head><title>Import in progress</title><meta name="color-scheme" content="light dark"></head>'
                        //                 + '<body><h2>Processing your import...</h2><button type=button onclick=window.close()>Close</button></body></html>';
                        //             popup.document.write(content);
                        //             return true; // let the browser continue form submission
                        //         }
                        //     },
                        //         m("select[name=importType][required].w-80", { title: "choose file format to export to" },
                        //             m("option", { value: "jsoncompact" }, ".json compact"),
                        //         ),
                        //         m("select[name=table]", { title: "choose table to import to" },
                        //             m("option", { value: "" }, "auto"),
                        //         ),
                        //         m("label.custom-file-label.w-80.no-wrap", {title: QryForm.selectedFileName || "choose file to upload"}, [
                        //             QryForm.selectedFileName || "file...",
                        //             m('input[type=file][name=file][required]', {
                        //                 style: { display: "none" },
                        //                 onchange: (e) => {
                        //                     QryForm.selectedFileName = e.target.files[0]?.name || "";
                        //                 }
                        //             })
                        //         ]),
                        //         m('input[name=conn][type="hidden"]', { value: App.conn }),
                        //         m('input[name=schema][type="hidden"]', { value: App.schema }),
                        //         m("button[type=submit].ml-10", {
                        //             title: "upload table and import data",
                        //         }, m(DownloadIcon)),
                        //     ),
                        // ),
                        m("div", { style: "float: right;" },
                            m("fieldset.ml-20",
                                m("legend", "query"),
                                m("button[type=button]", {
                                    disabled: QryForm.executing,
                                    onclick: () => {
                                        QryExplain.submit();
                                        App.tabState.set("explain");
                                    }
                                }, "explain"),
                                m("button[type=button].ml-10", {
                                    disabled: QryForm.executing,
                                    onclick: () => {
                                        QryForm.submitQuery();
                                        App.tabState.set("result");
                                    }
                                }, "execute"),
                                m("button[type=button]", {
                                    title: "abort execution",
                                    disabled: !QryForm.executing,
                                    onclick: () => {
                                        QryForm.xhr.abort();
                                        QryForm.xhr = null;
                                        QryForm.executing = false;
                                    }
                                }, "■"),
                            ),
                        )

                    ),
                ];
    }
}