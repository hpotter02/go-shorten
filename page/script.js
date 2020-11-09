function submitURL(e) {
    e.preventDefault();
    document.getElementById("errortext").innerText = "";
    let data = {
        destination: document.getElementById("urlInput").value,
    };
    fetch(".", {
        method: "POST",
        body: JSON.stringify(data)
    }).then(response => response.json())
        .then(result => {
            if (result.error != undefined) {
                document.getElementById("errortext").innerText = result.error;
            } else {
                document.getElementById("header").innerText = window.location.href + result.id;
            }

        })
        .catch(error => {
            console.error('Error:', error);
        });
}

function copyLink() {
    let text = document.getElementById("header").innerText;
    if (text != "go-shorten") {
        copyTextToClipboard(text);
    }

}

function copyTextToClipboard(text) {
    var textArea = document.createElement("textarea");
    textArea.style = { "display": "none" };
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        var successful = document.execCommand('copy');
        var msg = successful ? 'successful' : 'unsuccessful';

    } catch (err) {

    }

    document.body.removeChild(textArea);
}