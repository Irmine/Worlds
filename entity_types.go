package worlds

// EntityType is a type specifying the ID of an entity.
type EntityType int32

const (
	Chicken EntityType = iota + 10
	Pig
	Sheep
	Wolf
	Villager
	Mooshroom
	Squid
	Rabbit
	Bat
	IronGolem
	SnowGolem
	Ocelot
	Horse
	Donkey
	Mule
	SkeletonHorse
	ZombieHorse
	PolarBear
	Llama
	Parrot
)

const (
	Zombie EntityType = iota + 32
	Creeper
	Skeleton
	Spider
	ZombiePigman
	Slime
	Enderman
	SilverFish
	CaveSpider
	Ghast
	MagmaCube
	Blaze
	ZombieVillage
	Witch
	Stray
	Husk
	WitherSkeleton
	Guardian
	ElderGuardian
	NPC
	Wither
	EnderDragon
	Shulker
	Endermite
	LearnToCodeMascot
	Vindicator
)

const (
	ArmorStand EntityType = iota + 61
	TripodCamera
	Player
	Item
	Tnt
	FallingBlock
	MovingBlock
	XpBottle
	XpOrb
	EyeOfEnderSignal
	EnderCrystal
	FireworksRocket
)

const (
	ShulkerBullet EntityType = iota + 76
	FishingHook
	ChalkBoard
	DragonFireball
	Arrow
	Snowball
	Egg
	Painting
	Minecart
	LargeFireball
	SplashPotion
	EnderPearl
	LeashKnot
	WitherSkull
	Boat
	WitherSkullDangerous
	LightningBolt
	SmallFireball
	AreaEffectCloud
	HopperMinecart
	TntMinecart
	ChestMinecart
	Unused
	CommandBlockMinecart
	LingeringPotion
	LlamaSpit
	EvocationFang
	Evoker
	Vex
)
