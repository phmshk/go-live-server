const es = new EventSource("/live-reload");

es.onmessage = (event) => {
  const fileName = event.data;
  if (fileName.endsWith(".css")) {
    updateStylesOnPage();
  } else {
    window.location.reload();
  }
};

function updateStylesOnPage() {
  const linkTags = document.querySelectorAll('link[rel="stylesheet"]');
  for (let i = 0; i < linkTags.length; i++) {
    const link = linkTags[i];
    const url = new URL(link.href, window.location.href);
    url.searchParams.set("test_reload", Date.now());
    link.href = url.toString();
  }
}
