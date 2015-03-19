$(document).ready(function () {
    var highlights = {
        go: "go",
        python: "python",
        php: "php",
        xhr: "html",
        nodejs: "javascript",
        java: "java",
        objc: "objectivec",
        "objc.connection": "objectivec",
        vim: "vim"
    };
    function updateSnippet(event) {
        event.stopPropagation();
        event.preventDefault();
        var lang = $('#lang').val();
        var command = $('#command').val();
        if (command === "" || (command.indexOf("curl") === 0 && command.length < 6)) {
            return;
        }
        var result = CurlAsDsl.generate(lang, command);
        var codeBlock = $('#codeblock');
        for (var key in highlights) {
            if (highlights.hasOwnProperty(key)) {
                codeBlock.removeClass(highlights[key]);
            }
        }
        if (result[0]) {
            //document.getElementById('codeblock').innerHTML = result[0];
            codeBlock.html(result[0]);
            codeBlock.addClass(highlights[lang]);
            hljs.highlightBlock(document.getElementById('codeblock'));
        } else {
            codeBlock.text(result[1]);
        }
    }
    hljs.initHighlightingOnLoad();
    $('#lang').on('change', updateSnippet);
    $('#command').on('change', updateSnippet);
    $('#command').on('keypress', function (event) {
        if (event.which === 13) {
            updateSnippet(event);
        }
    });
    $('#button').on('click', updateSnippet);
    $('#reset').on('click', function () {
        $('#codeblock').html('Result code is printed here.');
    });
});
