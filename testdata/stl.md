# STL 3D Diagrams

## Pyramid

```stl
solid pyramid
  facet normal 0 0 -1
    outer loop
      vertex 0 0 0
      vertex 40 40 0
      vertex 40 0 0
    endloop
  endfacet
  facet normal 0 0 -1
    outer loop
      vertex 0 0 0
      vertex 0 40 0
      vertex 40 40 0
    endloop
  endfacet
  facet normal 0 -1 1
    outer loop
      vertex 0 0 0
      vertex 40 0 0
      vertex 20 20 35
    endloop
  endfacet
  facet normal 1 0 1
    outer loop
      vertex 40 0 0
      vertex 40 40 0
      vertex 20 20 35
    endloop
  endfacet
  facet normal 0 1 1
    outer loop
      vertex 40 40 0
      vertex 0 40 0
      vertex 20 20 35
    endloop
  endfacet
  facet normal -1 0 1
    outer loop
      vertex 0 40 0
      vertex 0 0 0
      vertex 20 20 35
    endloop
  endfacet
endsolid pyramid
```

## Wedge

```stl
solid wedge
  facet normal 0 0 -1
    outer loop
      vertex 0 0 0
      vertex 50 30 0
      vertex 50 0 0
    endloop
  endfacet
  facet normal 0 0 -1
    outer loop
      vertex 0 0 0
      vertex 0 30 0
      vertex 50 30 0
    endloop
  endfacet
  facet normal 0 -1 1
    outer loop
      vertex 0 0 0
      vertex 50 0 0
      vertex 50 0 25
    endloop
  endfacet
  facet normal 0 -1 1
    outer loop
      vertex 0 0 0
      vertex 50 0 25
      vertex 0 0 25
    endloop
  endfacet
  facet normal 1 0 0
    outer loop
      vertex 50 0 0
      vertex 50 30 0
      vertex 50 0 25
    endloop
  endfacet
  facet normal 0 1 0
    outer loop
      vertex 0 30 0
      vertex 0 0 25
      vertex 50 0 25
    endloop
  endfacet
  facet normal 0 1 0
    outer loop
      vertex 0 30 0
      vertex 50 0 25
      vertex 50 30 0
    endloop
  endfacet
  facet normal -1 0 0
    outer loop
      vertex 0 0 0
      vertex 0 0 25
      vertex 0 30 0
    endloop
  endfacet
endsolid wedge
```
