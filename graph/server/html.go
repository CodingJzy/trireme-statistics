package server

const js = `
<!DOCTYPE html>
<meta charset="utf-8">
<style>
    .link {
        stroke: #ccc;
    }

    #accept {
        fill: green;
    }

    .link.accept {
        stroke: green;
    }

    #reject {
        fill: red;
    }

    .link.reject {
        stroke: red;
    }

    #nowrejected {
        fill: orange;
    }

    .link.nowrejected {
        stroke: orange;
    }

    .node circle {
        fill: #696969;
        stroke: #fff;
        stroke-width: 1.5px;
    }

    .node text {
        pointer-events: none;
        font: 9px "Lucida Console", Monaco, monospace;
    }

    .namespace {
        width: 220px;
        border: 1px solid black;
        border-radius: 4px;
        margin-bottom: 1px;
    }

    .submit {
        border: 1px solid black;
        border-radius: 4px;
        font-size: 12px;
    }

    .endtime {
        margin-left: 19px;
        width: 220px;
        border: 1px solid black;
        border-radius: 4px;
    }

    .starttime {
        margin-left: 14px;
        width: 220px;
        border: 1px solid black;
        border-radius: 4px;
    }

    .set {
        text-align: center;
        font-family: sans-serif;
    }
</style>

<body>
    <form name="graphoptions" action="/graph">
        <div class="set">
            Start Time:
            <input name="starttime" class="starttime" type="datetime-local" step="1">
            <br> End Time:
            <input name="endtime" class="endtime" type="datetime-local" step="1">
            <br> Namespace:
            <input name="namespace" class="namespace" type="text">
            <br>
            <input type="submit" class="submit" value="Filter">
        </div>
    </form>
    <script src="//d3js.org/d3.v3.min.js"></script>
    <script>
        var width = 1000,
            height = 1000,
            radius = 8;
        var svg = d3.select("body").append("svg")
            .attr("viewBox", "0 0 " + width + " " + height)
            .attr("preserveAspectRatio", "xMidYMid meet")
            .attr("pointer-events", "all")
        var force = d3.layout.force()
            .gravity(0.05)
            .distance(200)
            .charge(-250)
            .size([850, 500]);
        d3.json({{.Address}}, function(error, json) {
            if (error) throw error;
            var edges = [];
            json.links.forEach(function(e) {
                var sourceNode = json.nodes.filter(function(n) {
                        return n.id === e.source;
                    })[0],
                    targetNode = json.nodes.filter(function(n) {
                        return n.id === e.target;
                    })[0];
                if (typeof sourceNode != "undefined" && typeof targetNode != "undefined") {
                    edges.push({
                        source: sourceNode,
                        target: targetNode,
                        time: e.time,
                        action: e.action,
                        namespace: e.namespace
                    });
                }
            });
            force
                .nodes(json.nodes)
                .links(edges)
                .on("tick", tick)
                .start();
            var marker = svg.append("svg:defs").selectAll("marker")
                .data(["accept", "reject", "nowrejected"])
                .enter().append("svg:marker")
                .attr("id", String)
                .attr("viewBox", "0 -5 10 10")
                .attr("refX", 0)
                .attr("refY", 0)
                .attr("markerWidth", 6)
                .attr("markerHeight", 6)
                .attr("orient", "auto")
                .append("svg:path")
                .attr("d", "M0,-5L10,0L0,5");
            var link = svg.selectAll(".link")
                .data(edges)
                .enter().append("polyline")
                .attr("class", function(d) {
                    return "link " + d.action;
                })
                .attr("marker-mid", function(d) {
                    return "url(#" + d.action + ")";
                });
            var node = svg.selectAll(".node")
                .data(json.nodes)
                .enter().append("g")
                .attr("class", "node")
                .on("mouseover", mouseover)
                .on("mouseout", mouseout)
                .call(force.drag);
            node.append("circle")
                .attr("r", radius);
            node.append("title")
                .text(function(d) {
                    return d.id;
                });
            node.append("text")
                .attr("dx", 10)
                .attr("dy", ".35em")
                .text(function(d) {
                    return d.name
                });

            var moveItems = (function() {
                var todoNode = 0;
                var todoLink = 0;
                var MAX_NODES = 300;
                var MAX_LINKS = MAX_NODES / 2;

                var restart = false;

                function moveSomeNodes() {
                    var n;
                    var goal = Math.min(todoNode + MAX_NODES, node[0].length);

                    for (var i = todoNode; i < goal; i++) {
                        n = node[0][i];
                        n.setAttribute("transform", "translate(" + Math.max(radius, Math.min(width - radius, n.__data__.x)) + "," +
                            Math.max(radius, Math.min(height - radius, n.__data__.y)) + ")");
                    }

                    todoNode = goal;
                    requestAnimationFrame(moveSome)
                }

                function moveSomeLinks() {
                    var l;
                    var goal = Math.min(todoLink + MAX_LINKS, link[0].length);

                    for (var i = todoLink; i < goal; i++) {
                        l = link[0][i];
                        l.setAttribute("points", l.__data__.source.x + "," + l.__data__.source.y + " " +
                            (l.__data__.source.x + l.__data__.target.x) / 2 + "," + (l.__data__.source.y + l.__data__.target.y) / 2 + " " +
                            l.__data__.target.x + "," + l.__data__.target.y);
                    }

                    todoLink = goal;
                    requestAnimationFrame(moveSome)
                }

                function moveSome() {
                    if (todoNode < node[0].length)
                        moveSomeNodes()
                    else {
                        if (todoLink < link[0].length)
                            moveSomeLinks()
                        else {
                            if (restart) {
                                restart = false;
                                todoNode = 0;
                                todoLink = 0;
                                requestAnimationFrame(moveSome);
                            }
                        }
                    }
                }

                return function moveItems() {
                    if (!restart) {
                        restart = true;
                        requestAnimationFrame(moveSome);
                    }
                };
            })();

            function tick() {
                moveItems();
            }

            function mouseover() {
                d3.select(this).select("circle").transition()
                    .duration(750)
                    .attr("r", 11);
            }

            function mouseout() {
                d3.select(this).select("circle").transition()
                    .duration(750)
                    .attr("r", radius);
            }
        });
    </script>
`
