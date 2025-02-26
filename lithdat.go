package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type SURFACE struct {
	Flags        uint32
	TextureIndex uint16
	TextureFlags uint16
}

type VEC2 struct {
	X float32
	Y float32
}

type VEC3 struct {
	X float32
	Y float32
	Z float32
}

type COLOR struct {
	R float32
	G float32
	B float32
}

type QUATERNION struct {
	X float32
	Y float32
	Z float32
	W float32
}

type TRIANGLE struct {
	T         [3]uint32
	PolyIndex uint32
}

type PLANE struct {
	Normal VEC3
	Dist   float32
}

type VERTEX struct {
	VPos      VEC3
	Uv0       VEC2
	Uv1       VEC2
	Color     int32
	VNormal   VEC3
	VTangent  VEC3
	VBinormal VEC3
}

type POLYGON struct {
	Plane   PLANE
	VertLen uint32
	VertPos []VEC3
}

type WMPOLYGON struct {
	SurfaceIndex     uint32
	PlaneIndex       uint32
	VerticesIndicies []uint32
}

type WMNODE struct {
	PolyIndex         uint32
	Zero              uint32   // unused leaves
	NodeSidesIndicies [2]int32 // children? // -1 == NODE_IN, -2 == NODE_OUT
}

type WORLDMODEL struct {
	Dummy             uint32
	WorldInfoFlags    uint32
	WorldName         string
	PointsLen         uint32
	PlanesLen         uint32
	SurfacesLen       uint32
	Portals           uint32
	PoliesLen         uint32
	LeavesLen         uint32
	PoliesVerticesLen uint32
	VisibleListLen    uint32
	LeafList          uint32
	NodesLen          uint32

	WorldBBoxMin     VEC3
	WorldBBoxMax     VEC3
	WorldTranslation VEC3

	TextureNamesSize uint32
	TextureNamesLen  uint32
	TextureNames     []string

	// number of verticesIndicies for polies
	VerticesLen []uint8
	Planes      []PLANE
	Surfaces    []SURFACE
	Polies      []WMPOLYGON

	Nodes []WMNODE

	Points []VEC3

	RootNodeIndex int32
	Sections      uint32 // unused (allways zero)
}

type WORLDTREE struct {
	RootBBoxMin    VEC3
	RootBBoxMax    VEC3
	SubNodesLen    uint32
	TerrainDepth   uint32
	WorldLayout    []uint8
	WorldModelsLen uint32
	WorldModels    []WORLDMODEL
}

type PROPERTY struct {
	Name           string
	DataTypeFlag   uint8 // 0: string, 1: VEC3, 2: COLOR, 3: Float, 5: uint8, 6: uint16, 7: QUATERNION
	PropertyFlags  uint32
	DataSize       uint16
	DataString     string
	DataVEC3       VEC3
	DataCOLOR      COLOR
	DataFloat      float32
	DataBool       bool
	DataUint32     uint32
	DataQuaternion QUATERNION
	Data           []byte
}

type WORLDOBJECT struct {
	ObjectSize    uint16
	ObjectType    string
	PropertiesLen uint32
	Properties    []PROPERTY
}

type LIGHTGRID struct {
	LookupStart VEC3
	BlockSize   VEC3
	LookupSize  [3]uint32
	LgDataLen   uint32
	LgData      []byte // RLE compressed data
}

type VERTICESPOS struct {
	Len  uint8
	Data []VEC3
}

type SKYPORTAL struct {
	VerticesPos VERTICESPOS
	Plane       PLANE
}

type OCCLUDER struct {
	VerticesPos VERTICESPOS
	Plane       PLANE
	Other       uint32
}

type SECTION struct {
	TextureName    []string
	ShaderCode     uint8
	TriangleCount  uint32
	TextureEffect  string
	LightMapWidth  uint32
	LightMapHeight uint32
	LightMapSize   uint32
	LightMapData   []byte // compressed
}

type SUBLM struct {
	Left   uint32
	Top    uint32
	Width  uint32
	Height uint32
	// RLE encoded data ubyte|ubyte|ubyte 0xFF|runCount|curData
	DataLen uint32
	Data    []byte
}

type SECTIONLM struct {
	SubLMLen uint32
	Sublm    []SUBLM
}

type LIGHTGROUP struct {
	Name   string
	VColor VEC3
	// zero compressed vertexIntensities
	// verticiesLen needed for decompressing
	NIntensitySize              uint32
	ZeroCompressedIntensityData []byte

	// section lightmap fix-ups
	SectionLMLen uint32
	SectionLM    []SECTIONLM
}

type RENDERNODE struct {
	VCenter           VEC3
	VHalfDims         VEC3
	SectionLen        uint32
	Sections          []SECTION
	VericesLen        uint32
	Vertices          []VERTEX
	TrianglesLen      uint32
	Triangles         []TRIANGLE
	SkyPortalLen      uint32
	Skyportals        []SKYPORTAL
	OccluderLen       uint32
	Occluders         []OCCLUDER
	LightGroupLen     uint32
	LightGroups       []LIGHTGROUP
	ChildFlags        uint8 // 0b(---- --11) last two digits show if children exist
	ChildNodeIndicies [2]uint32
}

type WMRENDERNODE struct {
	Name        string
	NodesLen    uint32
	Nodes       []RENDERNODE
	NoChildFlag uint32 // allways 0
}

type WORLDLIGHTGROUP struct {
	Name  string
	Color VEC3
	// poly data is saved as part of the rendering data
	VOffset VEC3
	VSize   VEC3
	Data    []byte
}

type RENDERDATA struct {
	// main world
	RenderTreeNodesLen uint32
	RenderNodes        []RENDERNODE
	// world models (adjusted for the physics and vis WM slots)
	WorldModelNodesLen    uint32
	WorldModelRenderNotes []WMRENDERNODE
	// light groups
	LightGroupsLen uint32
	LightGroups    []WORLDLIGHTGROUP
}

type HEADER struct {
	Version                uint32
	ObjectDataPos          uint32
	BlindObjectDataPos     uint32
	LightgridPos           uint32
	CollisionDataPos       uint32
	ParticleBlockerDataPos uint32
	RenderDataPos          uint32
	PackerType             uint32
	PackerVersion          uint32
	Future                 [6]uint32
}

type BLINDOBJECTDATA struct {
	Len    uint32
	DataID uint32
	Data   []byte
}

type WORLD struct {
	WorldInfoStrLen uint32
	WorldInfoStr    string
	WorldExtentsMin VEC3
	WorldExtentsMax VEC3
	WorldOffset     VEC3
	WorldTree       WORLDTREE
	// World objects
	WorldObjectsLen uint32
	WorldObjects    []WORLDOBJECT
	// Blind objects
	BlindObjectDataSize uint32
	BlindObjectsData    []BLINDOBJECTDATA
	// Lightgrid data
	LightGrid LIGHTGRID
	// Physics data
	PhysicsDataLen uint32
	Polies         []POLYGON
	Zero           uint32 // trailer for future expansion.
	// Particle blocker data
	ParticleBlockerDataLen uint32
	ParticleBlockers       []POLYGON
	Zero2                  uint32 // trailer for future expansion.
	// Render data
	RenderData RENDERDATA
}

// uses the leading 16 bits to determine the size of a string to read,
// reads into a buffer and then converts to a go string.
func readLithSTRING(file *os.File, destination *string) {
	var size uint16
	_ = binary.Read(file, binary.LittleEndian, &size)
	buf := make([]byte, size)
	_ = binary.Read(file, binary.LittleEndian, &buf)
	*destination = string(buf)
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
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Version)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.ObjectDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.BlindObjectDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.LightgridPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.CollisionDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.ParticleBlockerDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.RenderDataPos)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.PackerType)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.PackerVersion)
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[0])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[1])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[2])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[3])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[4])
	_ = binary.Read(worldFile, binary.LittleEndian, &worldHeader.Future[5])
	fmt.Println(worldHeader)

	var world WORLD

	// Parse world objects
	_, err = worldFile.Seek(int64(worldHeader.ObjectDataPos), io.SeekStart)
	fmt.Println("\nParsing WorldObjectsData @", worldHeader.ObjectDataPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.ObjectDataPos, err)
		os.Exit(1)
	}
	_ = binary.Read(worldFile, binary.LittleEndian, &world.WorldObjectsLen)
	fmt.Println("\nWorld Objects:", world.WorldObjectsLen)
	for range world.WorldObjectsLen {
		var worldObject WORLDOBJECT
		_ = binary.Read(worldFile, binary.LittleEndian, &worldObject.ObjectSize)
		readLithSTRING(worldFile, &worldObject.ObjectType)
		_ = binary.Read(worldFile, binary.LittleEndian, &worldObject.PropertiesLen)
		fmt.Println("\nProperty Count (", worldObject.ObjectType, "):", worldObject.PropertiesLen, "\n")
		for range worldObject.PropertiesLen {
			var property PROPERTY
			readLithSTRING(worldFile, &property.Name)
			_ = binary.Read(worldFile, binary.LittleEndian, &property.DataTypeFlag)
			_ = binary.Read(worldFile, binary.LittleEndian, &property.PropertyFlags)
			_ = binary.Read(worldFile, binary.LittleEndian, &property.DataSize)
			switch property.DataTypeFlag {
			case 0:
				readLithSTRING(worldFile, &property.DataString)
			case 1:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataVEC3)
			case 2:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataCOLOR)
			case 3:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataFloat)
			case 5:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataBool)
			case 6:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataUint32)
			case 7:
				_ = binary.Read(worldFile, binary.LittleEndian, &property.DataQuaternion)
			}
			fmt.Println(property)
			worldObject.Properties = append(worldObject.Properties, property)
		}
		world.WorldObjects = append(world.WorldObjects, worldObject)
	}

	_, err = worldFile.Seek(int64(worldHeader.BlindObjectDataPos), io.SeekStart)
	fmt.Println("\nParsing BlindObjectsData @", worldHeader.BlindObjectDataPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.BlindObjectDataPos, err)
		os.Exit(1)
	}
	_ = binary.Read(worldFile, binary.LittleEndian, &world.BlindObjectDataSize)
	fmt.Println(world.BlindObjectDataSize)

	_, err = worldFile.Seek(int64(worldHeader.LightgridPos), io.SeekStart)
	fmt.Println("\nParsing LightgridData @", worldHeader.LightgridPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.LightgridPos, err)
		os.Exit(1)
	}
	// _ = binary.Read(worldFile, binary.LittleEndian, &)
	// fmt.Println()

	_, err = worldFile.Seek(int64(worldHeader.CollisionDataPos), io.SeekStart)
	fmt.Println("\nParsing Physics/CollisionData @", worldHeader.CollisionDataPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.CollisionDataPos, err)
		os.Exit(1)
	}
	_ = binary.Read(worldFile, binary.LittleEndian, &world.PhysicsDataLen)
	fmt.Println(world.PhysicsDataLen)

	_, err = worldFile.Seek(int64(worldHeader.ParticleBlockerDataPos), io.SeekStart)
	fmt.Println("\nParsing ParticleBlockerData @", worldHeader.ParticleBlockerDataPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.ParticleBlockerDataPos, err)
		os.Exit(1)
	}
	_ = binary.Read(worldFile, binary.LittleEndian, &world.ParticleBlockerDataLen)
	fmt.Println(world.ParticleBlockerDataLen)

	_, err = worldFile.Seek(int64(worldHeader.RenderDataPos), io.SeekStart)
	fmt.Println("\nParsing RenderData @", worldHeader.ParticleBlockerDataPos)
	if err != nil {
		fmt.Printf("Error seeking to offset %d: %v\n", worldHeader.RenderDataPos, err)
		os.Exit(1)
	}

	// fmt.Println(world)
}
