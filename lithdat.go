package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

type SURFACE struct {
	flags        uint32
	textureIndex uint16
	textureFlags uint16
}

type VEC2 struct {
	x float32
	y float32
}

type VEC3 struct {
	x float32
	y float32
	z float32
}

type COLOR struct {
	r float32
	g float32
	b float32
}

type QUATERNION struct {
	x float32
	y float32
	z float32
	w float32
}

type TRIANGLE struct {
	t         [3]uint32
	polyIndex uint32
}

type PLANE struct {
	normal VEC3
	dist   float32
}

type VERTEX struct {
	vPos      VEC3
	uv0       VEC2
	uv1       VEC2
	color     int32
	vNormal   VEC3
	vTangent  VEC3
	vBinormal VEC3
}

type POLYGON struct {
	plane   PLANE
	vertLen uint32
	vertPos []VEC3
}

type WMPOLYGON struct {
	surfaceIndex     uint32
	planeIndex       uint32
	verticesIndicies []uint32
}

type WMNODE struct {
	polyIndex         uint32
	zero              uint32   // unused leaves
	nodeSidesIndicies [2]int32 // children? // -1 == NODE_IN, -2 == NODE_OUT
}

type WORLDMODEL struct {
	dummy             uint32
	worldInfoFlags    uint32
	worldName         string
	pointsLen         uint32
	planesLen         uint32
	surfacesLen       uint32
	portals           uint32
	poliesLen         uint32
	leavesLen         uint32
	poliesVerticesLen uint32
	visibleListLen    uint32
	leafList          uint32
	nodesLen          uint32

	worldBBoxMin     VEC3
	worldBBoxMax     VEC3
	worldTranslation VEC3

	textureNamesSize uint32
	textureNamesLen  uint32
	textureNames     []string

	// number of verticesIndicies for polies
	verticesLen []uint8
	planes      []PLANE
	surfaces    []SURFACE
	polies      []WMPOLYGON

	nodes []WMNODE

	points []VEC3

	rootNodeIndex int32
	sections      uint32 // unused (allways zero)
}

type WORLDTREE struct {
	rootBBoxMin    VEC3
	rootBBoxMax    VEC3
	subNodesLen    uint32
	terrainDepth   uint32
	worldLayout    []uint8
	worldModelsLen uint32
	worldModels    []WORLDMODEL
}

type PROPERTY struct {
	name           string
	dataTypeFlag   uint8 // 0: string, 1: VEC3, 2: COLOR, 3: Float, 5: uint8, 6: uint16, 7: QUATERNION
	propertyFlags  uint64
	dataSize       uint16
	dataString     string
	dataVEC3       VEC3
	dataCOLOR      COLOR
	dataFloat      float32
	dataUint8      uint8
	dataUint16     uint16
	dataQuaternion QUATERNION
	data           []byte
}

type WORLDOBJECT struct {
	objectSize    uint16
	objectType    string
	propertiesLen uint32
	properties    []PROPERTY
}

type LIGHTGRID struct {
	lookupStart VEC3
	blockSize   VEC3
	lookupSize  [3]uint32
	lgDataLen   uint32
	lgData      []byte // RLE compressed data
}

type VERTICESPOS struct {
	len  uint8
	data []VEC3
}

type SKYPORTAL struct {
	verticesPos VERTICESPOS
	plane       PLANE
}

type OCCLUDER struct {
	verticesPos VERTICESPOS
	plane       PLANE
	other       uint32
}

type SECTION struct {
	textureName    []string
	shaderCode     uint8
	triangleCount  uint32
	textureEffect  string
	lightMapWidth  uint32
	lightMapHeight uint32
	lightMapSize   uint32
	lightMapData   []byte // compressed
}

type SUBLM struct {
	left   uint32
	top    uint32
	width  uint32
	height uint32
	// RLE encoded data ubyte|ubyte|ubyte 0xFF|runCount|curData
	dataLen uint32
	data    []byte
}

type SECTIONLM struct {
	subLMLen uint32
	sublm    []SUBLM
}

type LIGHTGROUP struct {
	name   string
	vColor VEC3
	// zero compressed vertexIntensities
	// verticiesLen needed for decompressing
	nIntensitySize              uint32
	zeroCompressedIntensityData []byte

	// section lightmap fix-ups
	sectionLMLen uint32
	sectionlm    []SECTIONLM
}

type RENDERNODE struct {
	vCenter           VEC3
	vHalfDims         VEC3
	sectionLen        uint32
	sections          []SECTION
	vericesLen        uint32
	vertices          []VERTEX
	trianglesLen      uint32
	triangles         []TRIANGLE
	skyPortalLen      uint32
	skyportals        []SKYPORTAL
	occluderLen       uint32
	occluders         []OCCLUDER
	lightGroupLen     uint32
	lightGroups       []LIGHTGROUP
	childFlags        uint8 // 0b(---- --11) last two digits show if children exist
	childNodeIndicies [2]uint32
}

type WMRENDERNODE struct {
	name        string
	nodesLen    uint32
	nodes       []RENDERNODE
	noChildFlag uint32 // allways 0
}

type WORLDLIGHTGROUP struct {
	name  string
	color VEC3
	// poly data is saved as part of the rendering data
	vOffset VEC3
	vSize   VEC3
	data    []byte
}

type RENDERDATA struct {
	// main world
	renderTreeNodesLen uint32
	renderNodes        []RENDERNODE
	// world models (adjusted for the physics and vis WM slots)
	worldModelNodesLen    uint32
	worldModelRenderNotes []WMRENDERNODE
	// light groups
	lightGroupsLen uint32
	lightGroups    []WORLDLIGHTGROUP
}

type HEADER struct {
	version                uint32
	objectDataPos          uint32
	blindObjectDataPos     uint32
	lightgrid_pos          uint32
	collisionDataPos       uint32
	particleBlockerDataPos uint32
	renderDataPos          uint32
	packerType             uint32
	packerVersion          uint32
	future                 [6]uint32
}

type BLINDOBJECTDATA struct {
	len    uint32
	dataID uint32
	data   []byte
}

type WORLD struct {
	worldInfoStrLen uint32
	worldInfoStr    string
	worldExtentsMin VEC3
	worldExtentsMax VEC3
	worldOffset     VEC3
	worldTree       WORLDTREE
	// World objects
	worldObjectsLen uint32
	worldObjects    []WORLDOBJECT
	// Blind objects
	blindObjectDataSize uint32
	blindObjectsData    []BLINDOBJECTDATA
	// Lightgrid data
	lightGrid LIGHTGRID
	// Physics data
	physicsDataLen uint32
	polies         []POLYGON
	zero           uint32 // trailer for future expansion.
	// Particle blocker data
	particleBlockerDataLen uint32
	particleBlockers       []POLYGON
	zero2                  uint32 // trailer for future expansion.
	// Render data
	renderData RENDERDATA
}

func main() {
	filename := os.Args[1]
	worldFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer worldFile.Close()

	// Parse world header
	var worldHeader HEADER
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.version)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.objectDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.blindObjectDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.lightgrid_pos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.collisionDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.particleBlockerDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.renderDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.packerType)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.packerVersion)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[0])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[1])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[2])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[3])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[4])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.future[5])
	fmt.Println(worldHeader)

	// Parse WorldInfo

	// Parse WorldTree

	// DEBUG_BYTE

	// Parse WorldModelHeader

	// Parse Root WorldModel

	// Parse WorldModels

	// ->>>>  WorldBSP,

}
