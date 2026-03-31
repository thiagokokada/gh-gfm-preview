# GeoJSON and TopoJSON

## GeoJSON

```geojson
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "properties": {
        "name": "West"
      },
      "geometry": {
        "type": "Polygon",
        "coordinates": [
          [
            [-130, 20],
            [-100, 20],
            [-100, 50],
            [-130, 50],
            [-130, 20]
          ]
        ]
      }
    },
    {
      "type": "Feature",
      "properties": {
        "name": "East"
      },
      "geometry": {
        "type": "Polygon",
        "coordinates": [
          [
            [-100, 20],
            [-70, 20],
            [-70, 50],
            [-100, 50],
            [-100, 20]
          ]
        ]
      }
    }
  ]
}
```

## TopoJSON

```topojson
{
  "type": "Topology",
  "transform": {
    "scale": [1, 1],
    "translate": [0, 0]
  },
  "objects": {
    "regions": {
      "type": "GeometryCollection",
      "geometries": [
        {
          "type": "Polygon",
          "properties": {
            "name": "Left"
          },
          "arcs": [[0]]
        },
        {
          "type": "Polygon",
          "properties": {
            "name": "Right"
          },
          "arcs": [[1]]
        }
      ]
    }
  },
  "arcs": [
    [
      [0, 0],
      [40, 0],
      [0, 60],
      [-40, 0],
      [0, -60]
    ],
    [
      [50, 0],
      [40, 0],
      [0, 60],
      [-40, 0],
      [0, -60]
    ]
  ]
}
```
