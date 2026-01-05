CREATE OR REPLACE PROPERTY GRAPH InternetTopologyGraph
  NODE TABLES (
    AutonomousSystems KEY (ASN),
    Prefixes          KEY (PrefixCIDR),
    IXPs              KEY (IXPID),
    CollectorPeers    KEY (CollectorId, PeerId),
    RoutingProfiles   KEY (ProfileId)
  )
  EDGE TABLES (
    RoutingProfile_AS_Edges AS Transit
      SOURCE KEY (SourceASN) REFERENCES AutonomousSystems (ASN)
      DESTINATION KEY (TargetASN) REFERENCES AutonomousSystems (ASN),

    AS_Links AS ASRel
      SOURCE KEY (ASN) REFERENCES AutonomousSystems (ASN)
      DESTINATION KEY (TargetASN) REFERENCES AutonomousSystems (ASN),

    AS_IXP_Memberships AS MemberOf
      SOURCE KEY (ASN) REFERENCES AutonomousSystems (ASN)
      DESTINATION KEY (IXPID) REFERENCES IXPs (IXPID),

    RouteState AS HasRoute
      SOURCE KEY (CollectorId, PeerId) REFERENCES CollectorPeers (CollectorId, PeerId)
      DESTINATION KEY (PrefixCIDR) REFERENCES Prefixes (PrefixCIDR),

    RoutingProfile_Prefixes AS InScope
      SOURCE KEY (ProfileId) REFERENCES RoutingProfiles (ProfileId)
      DESTINATION KEY (PrefixCIDR) REFERENCES Prefixes (PrefixCIDR)
  );
