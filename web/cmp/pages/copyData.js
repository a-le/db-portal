function CopyDataPage() {
    return {
infoText: `This feature is still experimental.
- A new tab will open upon submission, either showing a report of the copied data or triggering a file download.
- There are no optimizations in the copy process to DB table; INSERT operations are performed per row.
`,
        origin: DataEndpointForm(),
        destination: DataEndpointForm(),
        getDestinationType: () => {
            const sel = document.querySelector('select[name="destination[type]"]');
            return sel && sel.value;
        },
        view: function () {
            // Standard HTML form for streamed file upload.
            return m("form", {
                method: "POST",
                action: "/api/copy",
                target: "exportpage", //this.getDestinationType() === "file" ? "exportpage" : "_self",
                enctype: "multipart/form-data",
                onsubmit: function (e) {
                    // Let the browser submit the form
                    return true;
                }
            }, [
                m("input[type=hidden][name=jwt]", { value: localStorage.getItem(JWT_KEY) }),

                m("fieldset.mb-20.w-600.h-130", { style: "display: block" },
                    m("legend", "Origin (copy from)"),
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
                m("pre", this.infoText)
            ]);
        }
    };
}
