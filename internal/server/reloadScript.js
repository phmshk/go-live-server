const es = new EventSource("/live-reload");

es.onmessage = (event) => {
  const fileName = event.data;
  if (fileName.endsWith(".css")) {
    updateStylesOnPage(fileName);
  } else {
    window.location.reload();
  }
};

function updateStylesOnPage(fileName) {
  const linkTags = document.querySelectorAll('link[rel="stylesheet"]');

  const normalizedServerFile = fileName.toLowerCase().replace(/\\/g, "/");

  for (let i = 0; i < linkTags.length; i++) {
    const link = linkTags[i];
    const url = new URL(link.href, window.location.href);

    const cleanPathname = url.pathname.toLowerCase();

    if (normalizedServerFile.endsWith(cleanPathname)) {
      url.searchParams.set("t", Date.now());
      link.href = url.toString();
    }
  }
}
