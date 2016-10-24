/**
 * Chapter 1
 * Creating a simple HTML5 full screen map
 *
 * Peter J Langley
 * http://www.codechewing.com
 */



$(document).ready(function() {

    var tiledRaster = new ol.layer.Tile({
        source: new ol.source.OSM()
    });
    var fieldStyle = new ol.style.Style({
        fill: new ol.style.Fill({
            color: 'rgba(47,195,96,0.6)'
        }),
        stroke: new ol.style.Stroke({
            color: 'white'
        })
    });

    var fieldJsonp = new ol.layer.Vector({
        source: new ol.source.Vector({
            url: 'geojson/output.json',
            format: new ol.format.GeoJSON()
        }),
        style: flowStyleFunction
    });

    var map = new ol.Map({
        view: new ol.View({
            center: ol.proj.transform([-81.6555, 43.614], 'EPSG:4326', 'EPSG:3857'),
            zoom: 13
        }),
        layers: [tiledRaster, fieldJsonp],
        controls: ol.control.defaults().extend([
            new ol.control.FullScreen()
        ]),
        target: 'js-map'
    });


    var Great = [53, 191, 0, 1];
    var Good = [115, 197, 0, 1];
    var Normal = [181, 203, 0, 1];
    var Slight = [210, 168, 0, 1];
    var Bad = [216, 170, 0, 1];
    var Severe = [229, 0, 26, 1];

    var SeverityLevel = {
        "Great": Great,
        "Good": Good,
        "Normal": Normal,
        "Slight": Slight,
        "Bad": Bad,
        "Severe": Severe
    };

    var defaultStyle = new ol.style.Style({
        fill: new ol.style.Fill({
            color: [250, 250, 250, 1]
        }),
        stroke: new ol.style.Stroke({
            color: [220, 220, 220, 1],
            width: 1
        })
    });

    var styleCache = {};

    function styleFunction(feature, resolution) {
        var properties = feature.getProperties();
        var level = feature.getProperties().sedimentlevel;
        if (!level || !SeverityLevel[level]) {
            return [defaultStyle];
        }
        if (!styleCache[level]) {
            styleCache[level] = new ol.style.Style({
                fill: new ol.style.Fill({
                    color: SeverityLevel[level]
                }),
                stroke: new ol.style.Stroke({
                    color: "white",
                    width: 1
                })
            });
        }
        return [styleCache[level]];
    }

    function flowStyleFunction(feature, resolution) {
        var properties = feature.getProperties();
        var level = feature.getProperties().flowlevel;
        if (!level || !SeverityLevel[level]) {
            return [defaultStyle];
        }
        if (!styleCache[level]) {
            styleCache[level] = new ol.style.Style({
                fill: new ol.style.Fill({
                    color: SeverityLevel[level]
                }),
                stroke: new ol.style.Stroke({
                    color: "white",
                    width: 1
                })
            });
        }
        return [styleCache[level]];
    };

    $("#control button").click(function(event) {
        var a = map.getLayers().getArray()[1];
        a.setStyle(styleFunction);
        alert(a);
    });
});
