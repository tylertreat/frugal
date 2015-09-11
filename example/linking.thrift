// Common fields in all API messages, both requests and responses
struct APIMessage {
    1: string AccountID, // Account resource identifier
    2: string MembershipID, // Membership resource identifier
    3: string UserID, // User resource identifier
    4: string CorrelationID // Correlation for action
}

// Used for each requested Atom update
struct AtomUpdate {
	1: i64 ID, // Atom ID
	2: string Value, // Value of the Atom
	3: string Target // Target location from which the Atom was last updated
}

// Message to request that Linking to update Atoms
struct AtomUpdateRequest {
	1: APIMessage Base, // Common part of all communications
	2: list<AtomUpdate> Updates // Atoms to update
}

// Message to request that Linking return current Atoms
struct GetCurrentAtomsRequest {
	1: APIMessage Base, // Common part of all communications
	2: list<string> AtomIDs // Atoms to get
}
