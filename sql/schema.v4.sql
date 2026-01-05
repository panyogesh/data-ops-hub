CREATE TABLE AutonomousSystems (
  ASN INT64 NOT NULL,
  FirstSeen TIMESTAMP,
  LastSeen  TIMESTAMP,
  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (ASN);

CREATE TABLE Prefixes (
  PrefixCIDR STRING(64) NOT NULL,
  IPVersion  INT64 NOT NULL,
  IsBogon BOOL,

  GeoCity STRING(MAX),
  GeoCountry STRING(MAX),
  ReputationScore FLOAT64,
  IsVPN BOOL,
  IsTorExit BOOL,
  DomainRegistrar STRING(MAX),
  LastEnrichedAt TIMESTAMP,

  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (PrefixCIDR);

CREATE TABLE IXPs (
  IXPID INT64 NOT NULL,
  Name STRING(MAX),
  City STRING(MAX),
  Country STRING(MAX),
  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (IXPID);

CREATE TABLE AS_IXP_Memberships (
  IXPID INT64 NOT NULL,
  ASN INT64 NOT NULL,
  FirstSeen TIMESTAMP,
  LastSeen  TIMESTAMP,
  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_Member_IXP FOREIGN KEY (IXPID) REFERENCES IXPs (IXPID),
  CONSTRAINT FK_Member_AS  FOREIGN KEY (ASN)   REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (IXPID, ASN),
  INTERLEAVE IN PARENT IXPs ON DELETE CASCADE;

CREATE INDEX AS_IXP_ByASN
ON AS_IXP_Memberships (ASN, IXPID);


CREATE TABLE Collectors (
  CollectorId STRING(64) NOT NULL,
  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (CollectorId);

CREATE TABLE CollectorPeers (
  CollectorId STRING(64) NOT NULL,
  PeerId      STRING(96) NOT NULL,
  PeerIP      STRING(64) NOT NULL,
  PeerASN     INT64,
  FirstSeen   TIMESTAMP,
  LastSeen    TIMESTAMP,
  UpdatedAt   TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_Peer_Collector FOREIGN KEY (CollectorId)
    REFERENCES Collectors (CollectorId),
  CONSTRAINT FK_Peer_AS FOREIGN KEY (PeerASN)
    REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (CollectorId, PeerId),
  INTERLEAVE IN PARENT Collectors ON DELETE CASCADE;

CREATE INDEX CollectorPeers_ByASN
ON CollectorPeers (PeerASN, CollectorId, PeerId);

CREATE TABLE MRT_Files (
  CollectorId   STRING(64) NOT NULL,
  FileId        STRING(96) NOT NULL,     
  FileType      STRING(16) NOT NULL,     
  FileTimestamp TIMESTAMP NOT NULL,      
  FileUri       STRING(MAX) NOT NULL,    

  ParsedAt   TIMESTAMP,
  IngestedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_File_Collector FOREIGN KEY (CollectorId)
    REFERENCES Collectors (CollectorId)
) PRIMARY KEY (CollectorId, FileId),
  INTERLEAVE IN PARENT Collectors ON DELETE CASCADE;

CREATE INDEX MRT_Files_ByTime
ON MRT_Files (CollectorId, FileType, FileTimestamp);


CREATE TABLE BGP_RIB_Routes (
  CollectorId STRING(64) NOT NULL,
  FileId      STRING(96) NOT NULL,

  PeerId      STRING(96) NOT NULL,
  PrefixCIDR  STRING(64) NOT NULL,
  ObservedAt  TIMESTAMP NOT NULL,     

  ASPath    ARRAY<INT64> NOT NULL,     
  OriginASN INT64 NOT NULL,            

  IngestedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_RIB_File FOREIGN KEY (CollectorId, FileId)
    REFERENCES MRT_Files (CollectorId, FileId),
  CONSTRAINT FK_RIB_Peer FOREIGN KEY (CollectorId, PeerId)
    REFERENCES CollectorPeers (CollectorId, PeerId),
  CONSTRAINT FK_RIB_Prefix FOREIGN KEY (PrefixCIDR)
    REFERENCES Prefixes (PrefixCIDR),
  CONSTRAINT FK_RIB_Origin FOREIGN KEY (OriginASN)
    REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (CollectorId, FileId, PeerId, PrefixCIDR),
  INTERLEAVE IN PARENT MRT_Files ON DELETE CASCADE;

CREATE INDEX BGP_RIB_ByPrefixTime
ON BGP_RIB_Routes (PrefixCIDR, ObservedAt);

CREATE INDEX BGP_RIB_ByOriginTime
ON BGP_RIB_Routes (OriginASN, ObservedAt);


CREATE TABLE BGP_UpdateEvents (
  CollectorId STRING(64) NOT NULL,
  FileId      STRING(96) NOT NULL,

  Shard      INT64 NOT NULL,          
  ObservedAt TIMESTAMP NOT NULL,       
  EventId    STRING(36) NOT NULL,      

  PeerId     STRING(96) NOT NULL,
  EventType  STRING(1) NOT NULL,       
  PrefixCIDR STRING(64) NOT NULL,

  ASPath    ARRAY<INT64>,
  OriginASN INT64,

  IngestedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_UPD_File FOREIGN KEY (CollectorId, FileId)
    REFERENCES MRT_Files (CollectorId, FileId),
  CONSTRAINT FK_UPD_Peer FOREIGN KEY (CollectorId, PeerId)
    REFERENCES CollectorPeers (CollectorId, PeerId),
  CONSTRAINT FK_UPD_Prefix FOREIGN KEY (PrefixCIDR)
    REFERENCES Prefixes (PrefixCIDR),
  CONSTRAINT FK_UPD_Origin FOREIGN KEY (OriginASN)
    REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (CollectorId, FileId, Shard, ObservedAt, EventId),
  INTERLEAVE IN PARENT MRT_Files ON DELETE CASCADE;

CREATE INDEX BGP_Updates_ByPrefixTime
ON BGP_UpdateEvents (PrefixCIDR, ObservedAt);

CREATE INDEX BGP_Updates_ByOriginTime
ON BGP_UpdateEvents (OriginASN, ObservedAt);

CREATE INDEX BGP_Updates_ByCollectorTime
ON BGP_UpdateEvents (CollectorId, ObservedAt);

CREATE INDEX BGP_Updates_ByPeerTime
ON BGP_UpdateEvents (CollectorId, PeerId, ObservedAt);


CREATE TABLE AS_Links (
  SourceASN INT64 NOT NULL,
  TargetASN INT64 NOT NULL,
  RelationshipType STRING(3) NOT NULL, -- 'c2p','p2c','p2p'
  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_ASL_Source FOREIGN KEY (SourceASN) REFERENCES AutonomousSystems (ASN),
  CONSTRAINT FK_ASL_Target FOREIGN KEY (TargetASN) REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (SourceASN, TargetASN),
  INTERLEAVE IN PARENT AutonomousSystems ON DELETE CASCADE;

CREATE INDEX AS_Links_ByTarget
ON AS_Links (TargetASN, ASN);

CREATE TABLE RouteState (
  CollectorId STRING(64) NOT NULL,
  PeerId      STRING(96) NOT NULL,
  PrefixCIDR  STRING(64) NOT NULL,

  IsActive        BOOL NOT NULL,
  LastObservedAt  TIMESTAMP NOT NULL, 
  LastEventType   STRING(1) NOT NULL,  

  CurrentASPath    ARRAY<INT64>,
  CurrentOriginASN INT64,

  UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true),

  CONSTRAINT FK_State_Peer FOREIGN KEY (CollectorId, PeerId)
    REFERENCES CollectorPeers (CollectorId, PeerId),
  CONSTRAINT FK_State_Prefix FOREIGN KEY (PrefixCIDR)
    REFERENCES Prefixes (PrefixCIDR),
  CONSTRAINT FK_State_Origin FOREIGN KEY (CurrentOriginASN)
    REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (CollectorId, PeerId, PrefixCIDR),
  INTERLEAVE IN PARENT CollectorPeers ON DELETE CASCADE;

CREATE INDEX RouteState_ByPrefix
ON RouteState (PrefixCIDR, CollectorId, PeerId);

CREATE INDEX RouteState_ByOriginTime
ON RouteState (CurrentOriginASN, LastObservedAt);


CREATE TABLE RoutingProfiles (
  ProfileId   STRING(36) NOT NULL,    
  ProfileType STRING(16) NOT NULL,    
  ScopeType   STRING(16) NOT NULL,    
  ScopeValue  STRING(MAX) NOT NULL,   
  WindowStart TIMESTAMP NOT NULL,
  WindowEnd   TIMESTAMP NOT NULL,
  CreatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
) PRIMARY KEY (ProfileId);

CREATE TABLE RoutingProfile_Prefixes (
  ProfileId  STRING(36) NOT NULL,
  PrefixCIDR STRING(64) NOT NULL,

  CONSTRAINT FK_RPP_Profile FOREIGN KEY (ProfileId)
    REFERENCES RoutingProfiles (ProfileId),
  CONSTRAINT FK_RPP_Prefix FOREIGN KEY (PrefixCIDR)
    REFERENCES Prefixes (PrefixCIDR)
) PRIMARY KEY (ProfileId, PrefixCIDR),
  INTERLEAVE IN PARENT RoutingProfiles ON DELETE CASCADE;

CREATE INDEX RoutingProfile_Prefixes_ByPrefix
ON RoutingProfile_Prefixes (PrefixCIDR, ProfileId);

CREATE TABLE RoutingProfile_AS_Edges (
  ProfileId STRING(36) NOT NULL,
  SourceASN INT64 NOT NULL,
  TargetASN INT64 NOT NULL,

  ObservedCount    INT64 NOT NULL,
  UniquePrefixes   INT64 NOT NULL,
  UniquePeers      INT64 NOT NULL,
  UniqueCollectors INT64 NOT NULL,

  FirstSeen TIMESTAMP NOT NULL,
  LastSeen  TIMESTAMP NOT NULL,

  CONSTRAINT FK_RPE_Profile FOREIGN KEY (ProfileId)
    REFERENCES RoutingProfiles (ProfileId),
  CONSTRAINT FK_RPE_Src FOREIGN KEY (SourceASN)
    REFERENCES AutonomousSystems (ASN),
  CONSTRAINT FK_RPE_Dst FOREIGN KEY (TargetASN)
    REFERENCES AutonomousSystems (ASN)
) PRIMARY KEY (ProfileId, SourceASN, TargetASN),
  INTERLEAVE IN PARENT RoutingProfiles ON DELETE CASCADE;

CREATE INDEX RoutingProfile_AS_Edges_ByTarget
ON RoutingProfile_AS_Edges (ProfileId, TargetASN, SourceASN);
