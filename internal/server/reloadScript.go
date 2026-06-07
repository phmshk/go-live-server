package server

const liveReloadScript = `
<script type="text/javascript">
(function() {
	if (!('EventSource' in window)) {
		console.error("Event Source is not found.");
		return;
	}
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
    	url.searchParams.set("t", Date.now());
    	link.href = url.toString();
  	}
	}

	updateStylesOnPage();
})();
</script>
`
