function CopyDataPage() {
    return {
        origin: DataEndpointForm(),
        destination: DataEndpointForm(),
        getDestinationType: () => {
            const sel = document.querySelector('select[name="destination[type]"]');
            return sel && sel.value;
        },
        view: function () {
            const self = this;
            // Standard HTML form for streamed file upload.
            return m("form", {
                method: "POST",
                action: "/api/copy",
                target: "exportpage", 
                enctype: "multipart/form-data",
                onsubmit: function (e) {
                    const popup = window.open('', 'exportpage', 'width=800,height=600');
                    const message = self.getDestinationType() === "file" 
                    ? "Preparing file. The download will start soon, please be patient..." 
                    : "Copying data. A report will be displayed, please be patient..."
                    if (popup) {
                        popup.document.write(`<html><head><style>body { color: #222; background: #fff; }@media (prefers-color-scheme: dark) {body { color: #eee; background: #222; }}</style></head><body><div>${message}</div><button onclick="window.close()">Close</button></body></html>`);
                        popup.document.close();
                    }
                    e.target.setAttribute('target', 'exportpage');
                    // Let the browser submit the form
                    return true;
                }
            }, [
                m("input[type=hidden][name=jwt]", { value: localStorage.getItem(JWT_KEY) }),

                m("fieldset.mb-20.w-600.h-130", { style: "display: block" },
                    m("legend", "Source (copy from)"),
                    m(this.origin, { endPointType: "origin" })
                ),
                m("fieldset.mb-20.w-600.h-130", { style: "display: block" },
                    m("legend", "Destination (copy to)"),
                    m(this.destination, { endPointType: "destination" })
                ),
                m("div.mb-20",
                    m("button[type=submit]", {
                        title: "copy data from origin to destination",
                        disabled: this.executing
                    }, this.getDestinationType() === "file" ? "download" : "copy data")
                ),
                //m("pre", this.infoText)
            ]);
        }
    };
}
